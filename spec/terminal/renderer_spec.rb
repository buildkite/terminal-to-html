# encoding: UTF-8

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
      long_string = "x" * 5 * 1024 * 1024
      last_part_of_long_string = render(long_string).split("").last(1000).join("")

      expect(last_part_of_long_string).to end_with("Warning: Terminal has chopped off the rest of the build as it&#39;s over the allowed 4 megabyte limit for logs.")
    end

    it "closes colors that get opened" do
      raw = "he\033[32mllo"

      expect(render(raw)).to eql("he<span class='term-fg32'>llo</span>")
    end

    it "skips over colors when backspacing" do
      raw = "he\e[32m\e[33m\bllo"

      expect(render(raw)).to eql("h<span class='term-fg33'>llo</span>")
    end

    it "treats \\e[39m a reset" do
      raw = "\e[36mthis has a color\e[39mthis is normal now\r\n"

      expect(render(raw)).to eql("<span class='term-fg36'>this has a color</span>this is normal now")
    end

    it "starts overwriting characters when you \\r midway through somehing" do
      raw = "hello\rb"

      expect(render(raw)).to eql("bello")
    end

    it "colors across multiple lines" do
      raw = "\e[32mhello\n\nfriend\e[0m"

      expect(render(raw)).to eql("<span class='term-fg32'>hello</span>\n&nbsp;\n<span class='term-fg32'>friend</span>")
    end

    it "allows you to control the cursor forwards" do
      raw = "this is\e[4Cpoop and stuff"

      expect(render(raw)).to eql("this is    poop and stuff")
    end

    it "doesn't allow you to jump down lines if the line doesn't exist" do
      raw = "this is great \e[1Bhello"

      expect(render(raw)).to eql("this is great hello")
    end

    it "allows you to control the cursor backwards" do
      raw = "this is good\e[4Dpoop and stuff"

      expect(render(raw)).to eql("this is poop and stuff")
    end

    it "allows you to control the cursor upwards" do
      raw = "1234\n56\e[1A78\e[B"

      expect(render(raw)).to eql("1278\n56")
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

      expect(render(raw)).to eql("aaaa\n1234\ncccc")
    end

    it "doesn't blow up if you go back too many characters" do
      raw = "this is good\e[100Dpoop and stuff"

      expect(render(raw)).to eql("poop and stuff")
    end

    it "\\e[1K clears everything before it" do
      raw = "hello\e[1Kfriend!"

      expect(render(raw)).to eql("     friend!")
    end

    it "clears everything after the \\e[0K" do
      raw = "hello\nfriend!\e[A\r\e[0K"

      expect(render(raw)).to eql("     \nfriend!")
    end

    it "handles \\e[0G ghetto style" do
      raw = "hello friend\e[Ggoodbye buddy!"

      expect(render(raw)).to eql("goodbye buddy!")
    end

    it "preserves characters already written in a certain color" do
      raw = "  \e[90m․\e[0m\e[90m․\e[0m\e[0G\e[90m․\e[0m\e[90m․\e[0m"

      expect(render(raw)).to eql("<span class='term-fgi90'>․․․․</span>")
    end

    it "replaces empty lines with non-breaking spaces" do
      raw = "hello\n\nfriend"

      expect(render(raw)).to eql("hello\n&nbsp;\nfriend")
    end

    it "preserves opening colors when using \\e[0G" do
      raw = "\e[33mhello\e[0m\e[33m\e[44m\e[0Ggoodbye"

      expect(render(raw)).to eql("<span class='term-fg33 term-bg44'>goodbye</span>")
    end

    it "allows erasing the current line up to a point" do
      raw = "hello friend\e[1K!"

      expect(render(raw)).to eql("            !")
    end

    it "allows clearing of the current line" do
      raw = "hello friend\e[2K!"

      expect(render(raw)).to eql("            !")
    end

    it "doesn't close spans if no colors have been opened" do
      raw = "hello \e[0mfriend"

      expect(render(raw)).to eql("hello friend")
    end

    it "\\e[K correctly clears all previous parts of the string" do
      raw = "remote: Compressing objects:   0% (1/3342)\e[K\rremote: Compressing objects:   1% (34/3342)"

      expect(render(raw)).to eql("remote: Compressing objects:   1% (34&#47;3342)")
    end

    it "collapses many spans of the same color into 1" do
      raw = "\e[90m․\e[90m․\e[90m․\e[90m․\n\e[90m․\e[90m․\e[90m․\e[90m․"

      expect(render(raw)).to eql("<span class='term-fgi90'>․․․․</span>\n<span class='term-fgi90'>․․․․</span>")
    end

    it "escapes HTML" do
      raw = "hello <strong>friend</strong>"

      expect(render(raw)).to eql("hello &lt;strong&gt;friend&lt;&#47;strong&gt;")
    end

    it "escapes HTML in color codes" do
      raw = "hello \e[\"hellomfriend"

      expect(render(raw)).to eql("hello \e[&quot;hellomfriend")
    end

    it "handles background colors" do
      raw = "\e[30;42m\e[2KOK (244 tests, 558 assertions)"

      expect(render(raw)).to eql("<span class='term-fg30 term-bg42'>OK (244 tests, 558 assertions)</span>")
    end

    it "handles xterm colors" do
      raw = "\e[38;5;169mhello\e[0m \e[38;5;179mgoodbye"

      expect(render(raw)).to eql("<span class='term-fgx169'>hello</span> <span class='term-fgx179'>goodbye</span>")
    end

    it "handles broken escape characters" do
      raw = "hi amazing \e[12 nom nom nom friends"

      expect(render(raw)).to eql("hi amazing \e[12 nom nom nom friends")
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

    it "handles colors with 3 attributes" do
      raw = "\e[0;10;4m\e[1m\e[34mgood news\e[0;10m\n\neveryone"

      expect(render(raw)).to eql("<span class='term-fg34 term-fg4 term-fg1'>good news</span>\n&nbsp;\neveryone")
    end

    it "ends underlining with \\e[24" do
      raw = "\e[4mbegin\e[24m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg4'>begin</span>\nend")
    end

    it "ends bold with \\e[21" do
      raw = "\e[1mbegin\e[21m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg1'>begin</span>\nend")
    end

    it "ends bold with \\e[22" do
      raw = "\e[1mbegin\e[22m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg1'>begin</span>\nend")
    end

    it "ends crossed out with \\e[29" do
      raw = "\e[9mbegin\e[29m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg9'>begin</span>\nend")
    end

    it "ends italic out with \\e[23" do
      raw = "\e[3mbegin\e[23m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg3'>begin</span>\nend")
    end

    it "ends decreased intensity with \\e[22" do
      raw = "\e[2mbegin\e[22m\r\nend"

      expect(render(raw)).to eql("<span class='term-fg2'>begin</span>\nend")
    end
  end

  private

  def render(raw)
    Terminal::Renderer.new(raw).render
  end
end
