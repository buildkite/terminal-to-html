require 'erb'

module Terminal
  class Preview
    class Binding
      def initialize(raw, rendered)
        @raw = raw
        @rendered = rendered
      end

      def raw
        # Call out special escape characters to make debugging easier
        @raw.
          gsub("\n", "\\n\n").
          gsub("\r", "\\r").
          gsub("\b", "\\b").
          gsub(/\e/, "\\\\e")
      end

      def rendered
        @rendered
      end

      def get_binding
        binding
      end
    end

    def initialize(raw, rendered)
      @raw = raw
      @rendered = rendered
    end

    def render
      template = File.read(template_path)
      renderer = ERB.new(template)
      binding = Binding.new(@raw, @rendered)

      renderer.result(binding.get_binding)
    end

    private

    def template_path
      File.join(File.expand_path(File.dirname(__FILE__)), 'templates/preview.html.erb')
    end
  end
end
