require 'escape_utils'

module Terminal
  class Renderer
    # If the string (which is a regex) matches, it'll split on that space, so we
    # end up with an array of special characters and normal characters, i.e:
    # [ '\n', '\r', 'a', 'b', '\e123m' ]
    SPLIT_BY_CHARACTERS = [
      # \n moves the cursor to a new line
      # \r moves the cursor to the begining of the line
      # \b moves the cursor back one
      '[\n\r\b]',

      # [K Erases from the current cursor position to the end of the current line.
      '\e\[0?K',

      # Clears tab at the current position
      '\e\[0?[Gg]',

      # [1K Erases from the current cursor position to the start of the current line.
      # [2K Erases the entire current line.
      '\e\[[1-2]K',

      # \e[?D move the cursor up ? many characters
      '\e\[[\d;]+A',

      # \e[?D move the cursor down ? many characters
      '\e\[[\d;]+B',

      # \e[?D move the cursor forward ? many characters
      '\e\[[\d;]+C',

      # \e[?D move the cursor back ? many characters
      '\e\[[\d;]+D',

      # \e[0m reset color information
      # \e[?m use the ? color going forward
      '\e\[[\d;]+m',

      # Random escpae sequences
      '\e',

      # Every other character
      '.'
    ]

    SPLIT_BY_CHARACTERS_REGEX = Regexp.new(SPLIT_BY_CHARACTERS.join("|"))

    def initialize(output)
      @output = output
      @screen = Screen.new
    end

    def render
      output = @output

      return "" if output.nil? || output.strip.length == 0

      # Limit the entire size of the output to 4 meg (4 * megabyte * kilabyte)
      max_total_size = 4 * 1024 * 1024
      if output.bytesize > max_total_size
        output = output.byteslice(0, max_total_size)
        output << "\n\nWarning: Terminal has chopped off the rest of the build as it's over the allowed 4 megabyte limit for logs."
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

    def emulate_terminal_rendering(string)
      # Splits the output into intersting parts.
      parts = string.scan(SPLIT_BY_CHARACTERS_REGEX)

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
          # Starts writing from a new line
          lines << line
          line = []
          cursor = 0
        when "\r"
          # Returns the writing cursor back to the begining of the line
          cursor = 0
        when "\e[G", "\e[0G", "\e[g"
          # TODO: I have no idea how these characters are supposed to work,
          # but this seems to be produce nicer results that what currently
          # gets rendered.
          cursor = 0
        when "\e[K", "\e[0K"
          # erases everything after the cursor
          line = line.fill(" ", cursor..line.length)
        when "\e[1K"
          # erases everything before the cursor
          line = line.fill(" ", 0..cursor)
        when "\e[2K"
          # erase entire line
          line = Array.new(line.length, " ")
        when "\b"
          pointer = cursor-1

          # Seek backwards until something that isn't a color is reached. When we
          # reach it (probably a string) remove the last character of it.
          # Colors aren't affected by \b
          while pointer >= 0
            char_at_pointer = line[pointer]

            unless char_at_pointer.kind_of?(Terminal::Node)
              line[pointer] = char_at_pointer[0..-2]
              break
            end

            pointer -= 1
          end
        when /\e\[(\d+)A/
          up = $1.to_i

          # TODO
        when /\e\[(\d+)B/
          down = $1.to_i

          # TODO
        when /\e\[(\d+)C/
          forwards = $1.to_i
          new_position = cursor + forwards

          # fill in the jumped spaces with a blank character
          line = line.fill(" ", line.length..(new_position - 1))

          cursor = new_position
        when /\e\[(\d+)D/
          backwards = $1.to_i

          new_position = cursor - backwards
          new_position = 0 if new_position < 0

          cursor = new_position
        when /\A\e\[(.*)m\z/
          color_code = $1.to_s

          # Determine what sort of color code it is.
          if color_code == "0"
            line[cursor] = Terminal::Reset.new(colors_opened)

            colors_opened = 0
          else
            line[cursor] = Terminal::Color.new(color_code)

            colors_opened += 1
          end

          index += char.length
          cursor += 1
        else
          line[cursor] = char
          cursor += 1
        end

        # The cursor can't go back furthur than 0, so if you \b at the begining
        # of a line, nothing happens.
        cursor = 0 if cursor < 0

        index += 1
      end

      # Be sure to reset any unclosed colors on the last line
      if colors_opened > 0
        line << Terminal::Reset.new(colors_opened)
      end

      # Add back in the last line if the end of the output
      # didn't end with a \n
      lines << line if line.any?

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
