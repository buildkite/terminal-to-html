# encoding: UTF-8

require 'escape_utils'
require 'strscan'

module Terminal
  class Renderer
    MEGABYTES = 1024 * 1024

    ESCAPE_CONTROL_CHARACTERS = "qQmKGgKAaBbCcDd".freeze
    ESCAPE_CAPTURE_REGEX = /\e\[(.*)([#{ESCAPE_CONTROL_CHARACTERS}])/

    INTERESTING_PARTS_REGEX=/[\n\r\b]|\e\[[\d;]*[#{ESCAPE_CONTROL_CHARACTERS}]|./

    def initialize(output, options = {})
      @output = output
      @options = options
      @screen = Screen.new
    end

    def render
      return "" if @output.nil?

      output = @output.dup

      render_to_screen(output)

      # Convert the screen to a string
      output = convert_screen_to_string

      # Escape any HTML
      escaped_html = escape_html(output)

      # Now convert the colors to HTML
      convert_to_html!(escaped_html)

      escaped_html
    end

    private

    def render_to_screen(string)
      scanner = StringScanner.new(string)

      # Scan the string for interesting things, for example:
      # [ '\n', '\r', 'a', 'b', '\e123m' ]
      while char = scanner.scan(INTERESTING_PARTS_REGEX)
        # The when cases are ordered by most likely, the lest checks it has to go
        # through before matching, the faster the render will be. Colors are
        # usually most likey, so that's first.
        if char == "\n".freeze
          @screen.x = 0
          @screen.y += 1
        elsif char == "\r".freeze
          @screen.x = 0
        elsif char == "\b".freeze
          @screen.x -= 1
        elsif char.index("\e".freeze) == 0 && char.length > 1
          sequence = char.match(ESCAPE_CAPTURE_REGEX)

          instruction = sequence[1]
          code = sequence[2]

          if code == "".freeze
            # no-op - an empty \e
          elsif code == "m".freeze
            @screen.color(instruction)
          elsif code == "G".freeze || code == "g".freeze
            @screen.x = 0
          elsif code == "K".freeze || code == "k".freeze
            if instruction == nil || instruction == "0".freeze
              # clear everything after the current x co-ordinate
              @screen.clear(@screen.y, @screen.x, Screen::END_OF_LINE)
            elsif instruction == "1".freeze
              # clear everything before the current x co-ordinate
              @screen.clear(@screen.y, Screen::START_OF_LINE, @screen.x)
            elsif instruction == "2".freeze
              @screen.clear(@screen.y, Screen::START_OF_LINE, Screen::END_OF_LINE)
            end
          elsif code == "A".freeze
            @screen.up(instruction)
          elsif code == "B".freeze
            @screen.down(instruction)
          elsif code == "C".freeze
            @screen.foward(instruction)
          elsif code == "D".freeze
            @screen.backward(instruction)
          end
        else
          @screen << char
        end
      end
    end

    def escape_html(string)
      EscapeUtils.escape_html(string)
    end

    def convert_screen_to_string
      @screen.to_s
    end

    def convert_to_html!(string)
      string.gsub!("\terminal[0]", "</span>")

      string.gsub!(/\terminal\[([^\]]+)\]/) do |match|
        %{<span class='#{$1}'>}
      end

      # Replace empty lines with a non breaking space.
      string.gsub!(/$^/, "&nbsp;")
    end
  end
end
