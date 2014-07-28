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
    EMPTY = " "
    END_OF_LINE = :end_of_line
    START_OF_LINE = :start_of_line

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
      line_length = line.length

      # Write empty slots until we reach the line
      (x - line_length).times do |i|
        line[line_length + i] = EMPTY
      end

      # Write the character to the slot
      line[x] = character
    end

    def <<(character)
      write(character)
      @x += 1
      character
    end

    def [](y)
      @screen[y]
    end

    def x=(value)
      @x = value > 0 ? value : 0
    end

    def y=(value)
      @y = value > 0 ? value : 0
    end

    def clear(y, x_start = nil, x_end = nil)
      if x_start.nil? && x_end.nil?
        @screen[y] = Array.new(@screen[y].length, EMPTY)
      else
        line = @screen[y]
        x_start = 0 if x_start == START_OF_LINE
        x_end = line.length if x_end == END_OF_LINE

        line.fill(EMPTY, x_start, x_end)
      end
    end

    def up(value = nil)
      self.y -= value.nil? ? 1 : value.to_i
    end

    def down(value = nil)
      increment = if value.nil?
                    1
                  else
                    value.to_i
                  end

      new_y = @y + increment

      # Only jump down if the line exists
      if @screen[new_y]
        self.y = new_y
      else
        false
      end
    end

    def backward(value = nil)
      self.x -= value.nil? ? 1 : value.to_i
    end

    def foward(value = nil)
      self.x += value.nil? ? 1 : value.to_i
    end

    def to_a
      @screen
    end

    def to_s
      @screen.to_a.map do |chars|
        chars.map(&:to_s).join("")
      end.join("\n")
    end
  end
end
