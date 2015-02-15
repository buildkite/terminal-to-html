# encoding: UTF-8

require 'escape_utils'
require 'emoji'

module Terminal
  class Renderer
    MEGABYTES = 1024 * 1024

    EMOJI_UNICODE_REGEXP = /[\u{1f600}-\u{1f64f}]|[\u{2702}-\u{27b0}]|[\u{1f680}-\u{1f6ff}]|[\u{24C2}-\u{1F251}]|[\u{1f300}-\u{1f5ff}]/
    EMOJI_IGNORE = [ "heavy_check_mark".freeze, "heavy_multiplication_x".freeze ]

    def initialize(output, options = {})
      @output = output

      @options = options
      @options[:emoji_asset_path] ||= "/assets/emojis"
      @options[:ansi2html_path] ||= "ansi2html"
    end

    def render
      return "" if @output.nil?

      # First duplicate the string, because we're going to be editing and chopping it
      # up directly.
      output = @output.to_s.dup

      # Don't allow parsing of outputs longer than 4 meg
      output = check_and_chomp_length(output)

      # Force encoding on the output first
      force_encoding!(output)

      # Call out to our Go program to convert to HTML
      output = ansi2html(output)

      # And emojify
      replace_unicode_with_emoji!(output)

      output
    end

    private

    def ansi2html output
      IO.popen(@options[:ansi2html_path], 'r+') do |p|
        p.write(output)
        p.close_write
        p.read
      end
    rescue Errno::ENOENT => e
      raise <<-EOF
Could not run ansi2html at '#{@options[:ansi2html_path]}'.
Install with `go install github.com/buildbox/terminal/cmd/ansi2html`.
Original exception: #{e}
      EOF
    end

    def check_and_chomp_length(output)
      # Limit the entire size of the output to 4 meg
      max_total_size = 4 * MEGABYTES
      if output.bytesize > max_total_size
        new_output = output.byteslice(0, max_total_size)
        new_output << "\n\nWarning: Terminal has chopped off the rest of the build as it's over the allowed 4 megabyte limit for logs."
        new_output
      else
        output
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
        path = File.join(@options[:emoji_asset_path], emoji.image_filename)

        %(<img alt="#{emoji.name}" title="#{emoji.name}" src="#{path}" class="emoji" width="20" height="20" />)
      else
        unicode
      end
    end
  end
end
