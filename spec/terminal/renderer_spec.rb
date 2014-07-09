require 'spec_helper'

describe Terminal::Renderer do
  let(:renderer) { Terminal::Renderer.new }

  describe "rendering of curl.sh" do
    it "returns the expected result" do
      fixture = Fixture.for("curl.sh")

      expect(renderer.render(fixture.raw)).to eql(fixture.rendered)
    end
  end

  describe "rendering of ascii.sh" do
    it "returns the expected result" do
      fixture = Fixture.for("ascii.sh")

      expect(renderer.render(fixture.raw)).to eql(fixture.rendered)
    end
  end

  describe "#render" do
    it "closes colors that get opened" do
      raw = "he\033[81mllo"

      expect(renderer.render(raw)).to eql("he<span class='c81'>llo</span>")
    end

    it "skips over colors when backspacing" do
      raw = "he\e[32m\e[33m\bllo"

      expect(renderer.render(raw)).to eql("h<span class='c32'><span class='c33'>llo</span></span>")
    end

    it "starts overwriting characters when you \\r midway through somehing" do
      raw = "hello\rb"

      expect(renderer.render(raw)).to eql("bello")
    end

    it "colors across multiple lines" do
      raw = "\e[81mhello\n\nfriend\e[0m"

      expect(renderer.render(raw)).to eql("<span class='c81'>hello\n&nbsp;\nfriend</span>")
    end
  end
end
