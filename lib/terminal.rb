require "terminal/version"
require "terminal/renderer"
require "terminal/cli"
require "terminal/preview"
require "terminal/engine" if defined?(Rails)

module Terminal
  def self.render(output, options = {})
    Terminal::Renderer.new(output, options).render
  end
end
