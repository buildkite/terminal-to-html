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
      '\e\[[\d;]*A',

      # \e[?D move the cursor down ? many characters
      '\e\[[\d;]*B',

      # \e[?D move the cursor forward ? many characters
      '\e\[[\d;]*C',

      # \e[?D move the cursor back ? many characters
      '\e\[[\d;]*D',

      # \e[0m reset color information
      # \e[?m use the ? color going forward
      '\e\[[\d;]+m',

      # Random escpae sequences
      '\e',

      # Every other character
      '.'
    ]

    SPLIT_BY_CHARACTERS_REGEX = Regexp.new(SPLIT_BY_CHARACTERS.join("|"))

    MEGABYTES = 1024 * 1024

    def initialize(output)
      @output = output
      @screen = Screen.new
    end

    def render
      return "" if @output.nil? || @output.strip.length == 0

      # First duplicate the string, because we're going to be editing and chopping it
      # up directly.
      output = @output.dup

      # Limit the entire size of the output to 4 meg
      max_total_size = 4 * MEGABYTES
      if output.bytesize > max_total_size
        output = output.byteslice(0, max_total_size)
        output << "\n\nWarning: Terminal has chopped off the rest of the build as it's over the allowed 4 megabyte limit for logs."
      end

      # Limit each line to (x) chars
      # TODO: Move this to the screen
      max_line_length = 50_000
      output = output.split("\n").map do |line|
        if line.length > max_line_length
          line = line[0..max_line_length]
          line << " Warning: Terminal has chopped the rest of this line off as it's over the allowed #{max_line_length} characters per line limit."
        else
          line
        end
      end.join("\n")

      # Force encoding on the output first
      force_encoding!(output)

      # Now do the terminal rendering (handles all the special characters)
      terminal_output = emulate_terminal_rendering(output)

      # Escape any HTML
      terminal_output = EscapeUtils.escape_html(terminal_output)

      # Now convert the colors to HTML
      colorize!(terminal_output)

      # Replace empty lines with a non breaking space.
      terminal_output.gsub(/$^/, "&nbsp;")
    end

    private

    def emulate_terminal_rendering(string)
      # Splits the output into intersting parts.
      parts = string.scan(SPLIT_BY_CHARACTERS_REGEX)

      colors_opened = 0

      # The when cases are ordered by most likely, the lest checks it has to go through
      # before matching, the faster the render will be. Colors are usually most likey, so that's first.
      parts.each_with_index do |char, index|
        case char
        when /\A\e\[(.*)m\z/
          code = $1.to_s
          escape = if code == "0"
                     # Only remove a color if we're > 1
                     colors_opened -= 1 if colors_opened > 0
                     Terminal::Reset.new
                   else
                     colors_opened += 1
                     Terminal::Color.new(code)
                   end

          @screen << escape
        when "\n"
          @screen.x = 0
          @screen.y += 1
        when "\r"
          @screen.x = 0
        when "\r"
          @screen.x = 0
        when "\b"
          # Seek backwards until something that isn't a color is reached. When
          # we reach it (probably a string) remove the last character of it.
          # Colors aren't affected by \b
          line = @screen[@screen.y]
          pointer = @screen.x - 1

          while pointer >= 0
            char_at_pointer = line[pointer]

            unless char_at_pointer.kind_of?(Terminal::Node)
              line[pointer] = ""
              break
            end

            pointer -= 1
          end
        when "\e[G", "\e[0G", "\e[g"
          # TODO: I have no idea how these characters are supposed to work,
          # but this seems to be produce nicer results that what currently
          # gets rendered.
          @screen.x = 0
        when "\e[K", "\e[0K"
          # clear everything after the current x co-ordinate
          @screen.clear(@screen.y, @screen.x, Screen::END_OF_LINE)
        when "\e[1K"
          # clear everything before the current x co-ordinate
          @screen.clear(@screen.y, Screen::START_OF_LINE, @screen.x)
        when "\e[2K"
          @screen.clear(@screen.y)
        when /\e\[(\d+)?A/
          @screen.up($1)
        when /\e\[(\d+)?B/
          @screen.down($1)
        when /\e\[(\d+)?C/
          @screen.foward($1)
        when /\e\[(\d+)?D/
          @screen.backward($1)
        else
          @screen << char
        end
      end

      colors_opened.times do
        @screen << Terminal::Reset.new
      end

      @screen.to_s
    end

    def colorize!(string)
      string.gsub!(/\e\[([0-9;]+)m/) do |match|
        color_codes = $1.split(';')

        if color_codes == [ "0" ]
          "</span>"
        else
          classes = color_codes.map { |code| "c#{code}" }

          "<span class='#{classes.join(" ")}'>"
        end
      end
    end

    def force_encoding!(string)
      string.force_encoding('UTF-8')

      if string.valid_encoding?
        string
      else
        string.force_encoding('ASCII-8BIT').encode!('UTF-8', invalid: :replace, undef: :replace)
      end
    end
  end
end
