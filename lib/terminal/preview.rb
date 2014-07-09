require 'erb'

module Terminal
  class Preview
    class Binding < Struct.new(:raw, :rendered)
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
