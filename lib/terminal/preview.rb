require 'erb'

module Terminal
  class Preview
    class Binding
      def initialize(asset_path, raw, rendered)
        @asset_path = asset_path
        @raw = raw
        @rendered = rendered
      end

      def asset_path(path)
        File.join(@asset_path, path)
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
      binding = Binding.new(assets_path, @raw, @rendered)

      renderer.result(binding.get_binding)
    end

    private

    def root_path
      File.expand_path(File.join(File.dirname(__FILE__), '..', '..'))
    end

    def assets_path
      File.join(root_path, 'app/assets')
    end

    def template_path
      File.join(root_path, 'lib/terminal/templates/preview.html.erb')
    end
  end
end
