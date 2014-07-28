require 'spec_helper'

describe Terminal::Screen do
  let(:screen) { Terminal::Screen.new }

  describe "#write" do
    it "writes to a given x/y co-ordinate" do
      screen.x = 3
      screen.y = 3
      screen.write('h')

      expect(screen.to_a).to eql([[], [], [], [" ", " ", " ", "h"]])
    end

    it "allows you to write to a co-ordinate via the method" do
      screen.write('a', 0, 0)
      screen.write('b', 1, 1)
      screen.write('c', 2, 2)

      expect(screen.to_a).to eql([["a"], [" ", "b"], [" ", " ", "c"]])

      expect(screen.x).to eql(0)
      expect(screen.y).to eql(0)
    end

    it "makes going back steps easy" do
      screen.x = 2
      screen.y = 2
      screen.write('b')
      screen.x -= 1
      screen.write('a')

      expect(screen.to_a).to eql([[], [], [" ", "a", "b"]])
    end

    it "sets the x to be 0 if you go into the negatives" do
      screen.x = -1

      expect(screen.x).to eql(0)
    end

    it "sets the y to be 0 if you go into the negatives" do
      screen.y = -1

      expect(screen.y).to eql(0)
    end
  end
end
