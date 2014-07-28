require 'spec_helper'

describe Terminal::Renderer do
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
      long_string = "x" * 4.5 * 1024 * 1024

      expect(render(long_string)).to end_with("Warning: Terminal has chopped the rest of this line off as it&#39;s over the allowed 50000 characters per line limit.</div>")
    end

    it "closes colors that get opened" do
      raw = "he\033[81mllo"

      expect(render(raw)).to eql("<div class='term-l'>he<span class='term-c81'>llo</span></div>")
    end

    it "skips over colors when backspacing" do
      raw = "he\e[32m\e[33m\bllo"

      expect(render(raw)).to eql("<div class='term-l'>h<span class='term-c32'><span class='term-c33'>llo</span></span></div>")
    end

    it "starts overwriting characters when you \\r midway through somehing" do
      raw = "hello\rb"

      expect(render(raw)).to eql("<div class='term-l'>bello</div>")
    end

    it "colors across multiple lines" do
      raw = "\e[81mhello\n\nfriend\e[0m"

      expect(render(raw)).to eql("<div class='term-l'><span class='term-c81'>hello</span></div><div class='term-l'><span class='term-c81'></span></div><div class='term-l'><span class='term-c81'>friend</span></div>")
    end

    it "allows you to control the cursor forwards" do
      raw = "this is\e[4Cpoop and stuff"

      expect(render(raw)).to eql("<div class='term-l'>this is    poop and stuff</div>")
    end

    it "doesn't allow you to jump down lines if the line doesn't exist" do
      raw = "this is great \e[1Bhello"

      expect(render(raw)).to eql("<div class='term-l'>this is great hello</div>")
    end

    it "allows you to control the cursor backwards" do
      raw = "this is good\e[4Dpoop and stuff"

      expect(render(raw)).to eql("<div class='term-l'>this is poop and stuff</div>")
    end

    it "allows you to control the cursor upwards" do
      raw = "1234\n56\e[1A78\e[B"

      expect(render(raw)).to eql("<div class='term-l'>1278</div><div class='term-l'>56</div>")
    end

    it "allows you to control the cursor downwards" do
      # creates a grid of:
      # aaaa
      # bbbb
      # cccc
      # Then goes up 2 rows, down 1 row, jumps to the begining
      # of the line, rewrites it to 1234, then jumps back down
      # to the end of the grid.
      raw = "aaaa\nbbbb\ncccc\e[2A\e[1B\r1234\e[1B"

      expect(render(raw)).to eql("<div class='term-l'>aaaa</div><div class='term-l'>1234</div><div class='term-l'>cccc</div>")
    end

    it "doesn't blow up if you go back too many characters" do
      raw = "this is good\e[100Dpoop and stuff"

      expect(render(raw)).to eql("<div class='term-l'>poop and stuff</div>")
    end

    it "\\e[1K clears everything before it" do
      raw = "hello\e[1Kfriend!"

      expect(render(raw)).to eql("<div class='term-l'>     friend!</div>")
    end

    it "clears everything after the \\e[0K" do
      raw = "hello\nfriend!\e[A\r\e[0K"

      expect(render(raw)).to eql("<div class='term-l'>     </div><div class='term-l'>friend!</div>")
    end

    it "handles \\e[0G ghetto style" do
      raw = "hello friend\e[Ggoodbye buddy!"

      expect(render(raw)).to eql("<div class='term-l'>goodbye buddy!</div>")
    end

    it "allows erasing the current line up to a point" do
      raw = "hello friend\e[1K!"

      expect(render(raw)).to eql("<div class='term-l'>            !</div>")
    end

    it "allows clearing of the current line" do
      raw = "hello friend\e[2K!"

      expect(render(raw)).to eql("<div class='term-l'>            !</div>")
    end
  end

  private

  def render(raw)
    Terminal::Renderer.new(raw).render
  end
end
