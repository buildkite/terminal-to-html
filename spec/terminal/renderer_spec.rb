require 'spec_helper'

describe Terminal::Renderer do
  let(:renderer) { Terminal::Renderer.new }

  describe "rendering of curl.sh" do
    it "returns the expected result" do
      fixture = Fixture.for("curl.sh")

      expect(renderer.render(fixture.raw)).to eql(fixture.rendered)
    end
  end
end
