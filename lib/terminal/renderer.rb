require 'escape_utils'

module Terminal
  class Renderer
    ESCAPE_CONTROL_CHARACTERS = "qQmKGgKAaBbCcDd"
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
      output = emulate_terminal_rendering(output)

      # Escape any HTML
      output = EscapeUtils.escape_html(output)

      # Now convert the colors to HTML
      convert_to_html(output)
    end

    private

    def force_encoding!(string)
      string.force_encoding('UTF-8')

      if string.valid_encoding?
        string
      else
        string.force_encoding('ASCII-8BIT').encode!('UTF-8', invalid: :replace, undef: :replace)
      end
    end

    def emulate_terminal_rendering(string)
      # Scan the string to create an array of interesting things, for example
      # it would look like this:
      # [ '\n', '\r', 'a', 'b', '\e123m' ]
      parts = string.scan(/[\n\r\b]|\e\[[\d;]*[#{ESCAPE_CONTROL_CHARACTERS}]|./)

      # The when cases are ordered by most likely, the lest checks it has to go
      # through before matching, the faster the render will be. Colors are
      # usually most likey, so that's first.
      parts.each do |char|
        # Hackers way of not having to run a regex over every
        # character.
        if char.length == 1
          case char
          when "\n"
            @screen.x = 0
            @screen.y += 1
          when "\r"
            @screen.x = 0
          when "\r"
            @screen.x = 0
          when "\b"
            @screen.x -= 1
          else
            @screen << char
          end
        else
          handle_escape_code(char)
        end
      end

      @screen.to_s
    end

    def handle_escape_code(sequence)
      # Escapes have the following: \e [ (instruction) (code)
      parts = sequence.match(/\e\[([\d;]+)?([#{ESCAPE_CONTROL_CHARACTERS}])/)

      instruction = parts[1].to_s
      code = parts[2].to_s

      case code
      when ""
        # no-op - an empty \e
      when "m"
        @screen.fg(instruction)
      when "G", "g"
        @screen.x = 0
      when "K", "k"
        case instruction
        when nil, "0"
          # clear everything after the current x co-ordinate
          @screen.clear(@screen.y, @screen.x, Screen::END_OF_LINE)
        when "1"
          # clear everything before the current x co-ordinate
          @screen.clear(@screen.y, Screen::START_OF_LINE, @screen.x)
        when "2"
          @screen.clear(@screen.y, Screen::START_OF_LINE, Screen::END_OF_LINE)
        end
      when "A"
        @screen.up(instruction)
      when "B"
        @screen.down(instruction)
      when "C"
        @screen.foward(instruction)
      when "D"
        @screen.backward(instruction)
      end
    end

    def convert_to_html(string)
      string = string.gsub(/\e\[((?:xfg|fg)?\d+)m/) do |match, x|
        if $1 == "0"
          %{</span>}
        else
          # the colors have already been sorted into xterm or regular colors
          # by the screen.
          %{<span class='term-#{$1}'>}
        end
      end

      # Replace empty lines with a non breaking space.
      string.gsub(/$^/, "&nbsp;")
    end
  end
end
