# encoding: UTF-8

require 'escape_utils'
require 'emoji'

module Terminal
  class Renderer
    MEGABYTES = 1024 * 1024

    EMOJI_UNICODE_REGEXP = /[\u{1f600}-\u{1f64f}]|[\u{2702}-\u{27b0}]|[\u{1f680}-\u{1f6ff}]|[\u{24C2}-\u{1F251}]|[\u{1f300}-\u{1f5ff}]/
    EMOJI_IGNORE = [ "heavy_check_mark", "heavy_multiplication_x" ]

    ESCAPE_CONTROL_CHARACTERS = "qQmKGgKAaBbCcDd"
    CAPTURE_ESCAPE_REGEX = /\e\[(.*)([#{ESCAPE_CONTROL_CHARACTERS}])/

    def initialize(output, options = {})
      @output = output

      @options = options
      @options[:emoji_asset_path] ||= "/assets/emojis"

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

      # Now do the render the output to the screen
      render_to_screen(output)

      # Convert the screen to a string
      output = convert_screen_to_string

      # Escape any HTML
      escaped_html = escape_html(output)

      # Now convert the colors to HTML
      convert_to_html!(escaped_html)

      # And emojify
      replace_unicode_with_emoji!(escaped_html)

      escaped_html
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

    def render_to_screen(string)
      # The when cases are ordered by most likely, the lest checks it has to go
      # through before matching, the faster the render will be. Colors are
      # usually most likey, so that's first.
      split_by_escape_character(string).each do |char|
        # Hackers way of not having to run a regex over every
        # character.
        if char.length == 1
          handle_character(char)
        else
          handle_escape_code(char)
        end
      end
    end

    def handle_character(char)
      if char == "\n"
        @screen.x = 0
        @screen.y += 1
      elsif char == "\r"
        @screen.x = 0
      elsif char == "\r"
        @screen.x = 0
      elsif char == "\b"
        @screen.x -= 1
      else
        @screen.write(char)
        @screen.x += 1
      end
    end

    def handle_escape_code(sequence)
      # Escapes have the following: \e [ (instruction) (code)
      parts = sequence.match(CAPTURE_ESCAPE_REGEX)

      instruction = parts[1].to_s
      code = parts[2].to_s

      if code == ""
        # no-op - an empty \e
      elsif code == "m"
        @screen.color(instruction)
      elsif code == "G" || code == "g"
        @screen.x = 0
      elsif code == "K" || code == "k"
        if instruction == nil || instruction == "0"
          # clear everything after the current x co-ordinate
          @screen.clear(@screen.y, @screen.x, Screen::END_OF_LINE)
        elsif instruction == "1"
          # clear everything before the current x co-ordinate
          @screen.clear(@screen.y, Screen::START_OF_LINE, @screen.x)
        elsif instruction == "2"
          @screen.clear(@screen.y, Screen::START_OF_LINE, Screen::END_OF_LINE)
        end
      elsif code == "A"|| code == "a"
        @screen.up(instruction)
      elsif code == "B"|| code == "b"
        @screen.down(instruction)
      elsif code == "C"|| code == "c"
        @screen.foward(instruction)
      elsif code == "D"|| code == "d"
        @screen.backward(instruction)
      end
    end

    def escape_html(string)
      EscapeUtils.escape_html(string)
    end

    def convert_screen_to_string
      @screen.to_s
    end

    # Scan the string to create an array of interesting things, for example
    # it would look like this:
    # [ '\n', '\r', 'a', 'b', '\e123m' ]
    def split_by_escape_character(string)
      string.scan(/[\n\r\b]|\e\[[\d;]*[#{ESCAPE_CONTROL_CHARACTERS}]|./)
    end

    def convert_to_html!(string)
      string.gsub!("\terminal[0]", "</span>")

      string.gsub!(/\terminal\[([^\]]+)\]/) do |match|
        %{<span class='#{$1}'>}
      end

      # Replace empty lines with a non breaking space.
      string.gsub!(/$^/, "&nbsp;")
    end

    def replace_unicode_with_emoji!(string)
      string.gsub!(EMOJI_UNICODE_REGEXP) do |match|
        emoji_image_from_unicode(match)
      end
    end

    # The Emoji API will be transitioning to a nil-based find API, at the
    # moment it raies exceptions for Emojis that can't be found:
    # https://github.com/github/gemoji/commit/b1736a387c7c1c2af300506fea5603e2e1fb89d8
    # Will support both for now.
    def emoji_image_from_unicode(unicode)
      emoji = Emoji.find_by_unicode(unicode)

      if emoji && !EMOJI_IGNORE.include?(emoji.name)
        name = ":#{emoji.name}:"
        path = File.join(@options[:emoji_asset_path], emoji.image_filename)

        %(<img alt="#{name}" title="#{name}" src="#{path}" class="emoji" width="20" height="20" />)
      else
        unicode
      end
    rescue
      unicode
    end
  end
end
