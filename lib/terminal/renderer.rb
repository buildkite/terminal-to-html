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

      # \e[0m reset color information
      # \e[?m use the ? color going forward
      '\e\[[\d;]+m',

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

      # Random escpae sequences
      '\e',

      # Every other character
      '.'
    ]

    SPLIT_BY_CHARACTERS_REGEX = Regexp.new(SPLIT_BY_CHARACTERS.join("|"))
    COLOR_REGEX = /\e\[.+m/
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
      convert_to_html(terminal_output)
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
      # Splits the output into intersting parts.
      parts = string.scan(SPLIT_BY_CHARACTERS_REGEX)

      # The when cases are ordered by most likely, the lest checks it has to go through
      # before matching, the faster the render will be. Colors are usually most likey, so that's first.
      parts.each do |char|
        case char
        when "\n"
          @screen.x = 0
          @screen.y += 1
        when "\r"
          @screen.x = 0
        when "\r"
          @screen.x = 0
        when "\b"
          # Colors are skipped when backspacing
          line = @screen[@screen.y]
          pointer = @screen.x

          while c = line[pointer -= 1]
            if not c =~ COLOR_REGEX # color
              line[pointer] = ""
              break
            end
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

      @screen.to_s
    end

    def convert_to_html(string)
      buffer = [ "<div class='term-l'>" ]

      # An array of open spans. As spans get opened, they get
      # added to this list, and as they're reset (\e[0m) they are
      # removed.
      opened_spans = []

      # Split the string into colors, new lines and the rest
      chunks = string.scan(/\e\[[\d;]+m|\n|[^\n\e]+/)

      # Iterate over the chunks, and turn them into HTML
      chunks.each do |chunk|
        case chunk
        when "\n"
          # Close any open spans
          opened_spans.length.times do
            buffer << "</span>"
          end

          # End the line and start a new one
          buffer << "</div><div class='term-l'>"

          # If there were open spans on the previous line, open them up again
          # here.
          opened_spans.each do |span|
            buffer << span
          end
        when /\e\[(.*)m/
          # Extract the color codes
          code = $1.to_s

          if code == "0"
            # Only close a span if one has been opened. So if you have a string lin
            # hello\e[0mthere it doens't just put a close span in there for no reason.
            if opened_spans.length > 0
              opened_spans.shift
              buffer << "</span>"
            end
          else
            codes = code.split(";")

            # Figure out what class name to use. Supports xterm256 colors
            # indexes.
            class_name = if codes[0] == "38" && codes[1] == "5"
                           "term-fg#{codes[2]}"
                         elsif codes[0] == "48" && codes[1] == "5"
                           raise 'yolo'
                           "term-bg#{codes[2]}"
                         elsif codes.length == 1
                           "term-fg#{codes.last}"
                         else
                           "term-esc-unknown"
                         end

            span = "<span class='#{class_name}'>"

            # Add this span to the list of opened ones
            opened_spans << span

            # Open a span that is the color
            buffer << span
          end
        else
          # If it's not a new color, or a new line, then add the chunk
          # to the buffer.
          buffer << chunk
        end
      end

      # If there were any open spans left over, close them
      opened_spans.length.times do
        buffer << "</span>"
      end

      # End off the final line
      buffer << "</div>"

      buffer.join("")
    end
  end
end
