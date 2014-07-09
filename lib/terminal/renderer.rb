require 'escape_utils'

module Terminal
  class Renderer
    def self.render(string)
      new.render(output)
    end

    def render(output)
      return "" if output.nil?

      # Limit the entire size of the output to 4 meg
      max_total_size = 4.megabytes
      if output.bytesize > max_total_size
        output = output.byteslice(0, max_total_size)
        output << "\n\nWarning: Terminal has chopped off the rest of the build as it's over the allowed #{number_to_human_size(max_total_size)} limit for logs."
      end

      # Limit each line to (x) chars
      max_line_length = 50_000
      output = output.split("\n").map do |line|
        if line.length > max_line_length
          line = line[0..max_line_length]
          line << " Warning: Terminal has chopped the rest of this line off as it's over the allowed #{max_line_length} characters per line limit."
        else
          line
        end
      end.join("\n")

      # Now do the terminal rendering
      output = emulate_terminal_rendering(sanitize(output))

      # Replace empty lines with a non breaking space.
      output.gsub(/$^/, "&nbsp;")
    end

    private

    # \n moves the cursor to a new line
    # \r moves the cursor to the begining of the line
    # \b moves the cursor back one
    # \e[0m reset color information
    # \e[?m use the ? color going forward
    def emulate_terminal_rendering(string)
      return "" if string.blank?

      # Splits the output into intersting parts.
      parts = string.scan(/[\n\r\b]|\e\[[\d;]+m|\e|[^\n\r\b\e]+/)

      lines = []

      index = 0
      length = string.length

      line = []
      cursor = 0

      # Every time a color is found, we increment this
      # counter, every time it resets, we decrement.
      # We do this so we close all the open spans at the
      # end of the output.
      colors_opened = 0

      parts.each do |char|
        case char
        when "\n"
          lines << line
          line = []
          cursor = 0
        when "\r"
          cursor = 0
        when "\b"
          pointer = cursor-1

          # Seek backwards until something that isn't a color is reached.
          # Colors aren't affected by \b
          while pointer > 0
            char_at_pointer = line[pointer]
            break unless char_at_pointer.kind_of?(Terminal::Color)

            pointer -= 1
          end

          cursor -= (cursor - pointer)
        when "\e"
          # Seek the next few characters to tell if theres a color present
          seeked = string[index..(index + 10)]

          # Does the next lot of characters look like a color code?
          matched = seeked.to_s.match(/\A\e\[(.*)m\z/)

          # If it does, skip over the characters that are the color,
          # and track it at a single color object in the line. We do this
          # so when we \b after a color, we skip back across the whole color
          # not just the last character in the color sequence.
          if matched
            color_code = matched[1].to_s

            # Determine what sort of color code it is.
            if color_code == "0"
              line[cursor] = Terminal::Reset.new(colors_opened)

              colors_opened = 0
            else
              line[cursor] = Terminal::Color.new(color_code)

              colors_opened += 1
            end

            index += 3 + color_code.length
            cursor += 1

            next
          end
        else
          line[cursor] = char
          cursor += 1
        end

        # The cursor can't go back furthur than 0, so if you \b at the begining
        # of a line, nothing happens.
        cursor = 0 if cursor < 0

        index += 1
      end

      # Add back in the last line if the end of the output
      # didn't end with a \n
      lines << line if line.any?

      # Be sure to reset any unclosed colors.
      if colors_opened > 0
        lines << [ Terminal::Reset.new(colors_opened) ]
      end

      # Join all the strings back together again
      lines = lines.map do |parts|
        completed_line = parts.map(&:to_s).join("")
      end.join("\n")

      # Now escape all the things
      lines = EscapeUtils.escape_html(lines)

      matches = []

      # Now we can easily gsub colors like a baws
      lines.gsub!(/\e\[([0-9;]+)m/) do |match|
        color_codes = $1.split(';')

        if color_codes == [ "0" ]
          "</span>"
        else
          classes = color_codes.map { |code| "c#{code}" }

          "<span class='#{classes.join(" ")}'>"
        end
      end

      lines
    end

    def sanitize(string)
      string = string.dup.force_encoding('UTF-8')
      if string.valid_encoding?
        string
      else
        string.
          force_encoding('ASCII-8BIT').
          encode!('UTF-8',
                  invalid: :replace,
                  undef:   :replace)
      end
    end
  end
end
