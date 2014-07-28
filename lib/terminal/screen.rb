# A screen with x/y co-ordinates.
#
# screen = Screen.new
# screen.x = 10
# screen.y = 10
#
# Will essentially create a 10 x 10 matrix with empty characters.
#
# screen.x = 5
# screen.y = 5
# screen.write 'y'
#
# Will write the 'y' character to the 5,5 slot.
#
# screen.to_a returns an array of the screen.
#
# Co-ordinates start at 0,0

module Terminal
  class Screen
    attr_reader :x, :y

    def initialize
      @x = 0
      @y = 0
      @screen = []
    end

    def write(character, x = @x, y = @y)
      # Expand the screen if we need to
      ((y + 1) - @screen.length).times do
        @screen << []
      end

      line = @screen[y]

      # Write empty slots until we
      (x - line.length).times do |i|
        line[i] = " "
      end

      # Write the character to the slot
      line[x] = character
    end

    def x=(value)
      @x = value > 0 ? value : 0
    end

    def y=(value)
      @y = value > 0 ? value : 0
    end

    def to_a
      @screen
    end
  end
end
