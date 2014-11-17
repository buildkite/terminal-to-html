# encoding: UTF-8

require 'spec_helper'

describe Terminal::Renderer do
  before :all do
    system "go build cmd/ansi2html/ansi2html.go"
    if $?.to_i != 0
      raise "Could not build ansi2html binary, can't run tests"
    end
  end

  describe "rendering of curl.sh" do
    it "returns the expected result" do
      fixture = Fixture.for("curl.sh")

      expect(render(fixture.raw)).to eql(fixture.rendered)
    end
  end

  describe "rendering of homer.sh" do
    it "returns the expected result" do
      fixture = Fixture.for("homer.sh")

      expect(render(fixture.raw)).to eql(fixture.rendered)
    end
  end

  describe "#render" do
    it "chops off logs longer than 4 megabytes" do
      long_string = "x" * 5 * 1024 * 1024
      last_part_of_long_string = render(long_string).split("").last(1000).join("")

      expect(last_part_of_long_string).to end_with("Warning: Terminal has chopped off the rest of the build as it&#39;s over the allowed 4 megabyte limit for logs.")
    end

    it "renders unicode emoji" do
      raw = "this is great 👍"

      expect(render(raw)).to eql(%{this is great <img alt=":+1:" title=":+1:" src="/assets/emojis/unicode/1f44d.png" class="emoji" width="20" height="20" />})
    end

    it "returns nothing if the unicode emoji can't be found" do
      expect(Emoji).to receive(:unicodes_index) { {} }
      raw = "this is great 😎"

      expect(render(raw)).to eql(%{this is great 😎})
    end

    it "leaves the tick emoji alone (it looks better and is colored)" do
      raw = "works ✔"

      expect(render(raw)).to eql(%{works ✔})
    end

    it "leaves the ✖ emoji alone as well" do
      raw = "broke ✖"

      expect(render(raw)).to eql(%{broke ✖})
    end
  end

  private

  def render(raw)
    Terminal::Renderer.new(raw, ansi2html_path: './ansi2html').render
  end
end
