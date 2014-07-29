# encoding: UTF-8

require 'spec_helper'

describe Terminal::Screen do
  let(:screen) { Terminal::Screen.new }

  describe "#write" do
    it "writes to a given x/y co-ordinate" do
      screen.x = 0
      screen.y = 0
      screen.write('a')

      screen.x = 1
      screen.y = 1
      screen.write('b')

      screen.x = 2
      screen.y = 2
      screen.write('c')

      expect(screen.to_a).to eql([["a"], [" ", "b"], [" ", " ", "c"]])
    end

    it "makes going back steps easy" do
      screen.x = 2
      screen.y = 2
      screen.write('b')
      screen.x -= 1
      screen.write('a')

      expect(screen.to_a).to eql([[], [], [" ", "a", "b"]])
    end

    it "makes going foward steps easy" do
      screen.x = 3
      screen.write('a')
      screen.x = 7
      screen.write('b')

      expect(screen.to_a).to eql([[" ", " ", " ", "a", " ", " ", " ", "b"]])
    end

    it "sets the x to be 0 if you go into the negatives" do
      screen.x = -1

      expect(screen.x).to eql(0)
    end

    it "sets the y to be 0 if you go into the negatives" do
      screen.y = -1

      expect(screen.y).to eql(0)
    end

    it "gives you a shortcut function to add to the current line" do
      screen << 'h'
      screen << 'i'

      expect(screen.to_a).to eql([["h", "i"]])
    end

    it "can convert the screen to a string" do
      screen << 'h'
      screen << 'i'
      screen.x = 0
      screen.y += 1
      screen << 'there'

      expect(screen.to_s).to eql("hi\nthere")
    end

    it "does nothing if trying to clear a line that doesn't exist" do
      screen.clear(12, Terminal::Screen::START_OF_LINE, Terminal::Screen::END_OF_LINE)

      expect(screen.to_s).to eql("")
    end
  end
end
