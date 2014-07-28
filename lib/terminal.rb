require "terminal/version"
require "terminal/node"
require "terminal/color"
require "terminal/reset"
require "terminal/screen"
require "terminal/renderer"
require "terminal/cli"
require "terminal/preview"

module Terminal
  def self.render(output)
    Terminal::Renderer.new.render(output)
  end
end
