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
    class Node < Struct.new(:blob, :fg, :bg)
      def ==(value)
        blob == value
      end

      def to_s
        blob
      end
    end

    END_OF_LINE = :end_of_line
    START_OF_LINE = :start_of_line
    EMPTY = Node.new(" ")

    attr_reader :x, :y

    def initialize
      @x = 0
      @y = 0
      @screen = []
      @fg = nil
    end

    def write(character)
      # Expand the screen if we need to
      ((@y + 1) - @screen.length).times do
        @screen << []
      end

      line = @screen[@y]
      line_length = line.length

      # Write empty slots until we reach the line
      (@x - line_length).times do |i|
        line[line_length + i] = EMPTY
      end

      # Write the character to the slot
      line[@x] = Node.new(character, @fg, @bg)
    end

    def <<(character)
      write(character)
      @x += 1
      character
    end

    def x=(value)
      @x = value > 0 ? value : 0
    end

    def y=(value)
      @y = value > 0 ? value : 0
    end

    def clear(y, x_start, x_end)
      line = @screen[y]

      x_start = 0 if x_start == START_OF_LINE
      x_end = line.length - 1 if x_end == END_OF_LINE

      if x_start == START_OF_LINE && x_end == END_OF_LINE
        @screen[y] = []
      else
        line.fill(EMPTY, x_start..x_end)
      end
    end

    # Changes the current foreground color that all new characters
    # will be written with.
    def fg(color)
      if color == "0" # reset all styles
        @fg = nil
      elsif color == "39" # reset text style
        @fg = nil
      else
        @fg = color
      end

      color
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
      @screen.to_a.map { |chars| chars.map(&:to_s) }
    end

    # Renders each node to a string, inserting and cleaning up color escape
    # sequences where neccessary.
    def to_s
      last_line_index = @screen.length - 1
      buffer = []

      @screen.each_with_index do |line, line_index|
        previous = nil
        open_fgs = 0

        line.each do |node|
          # If there is no previous node, and the current node has a color
          # (think first node in a line) then add the escape character.
          if !previous && node.fg
            buffer << "\e[#{node.fg}m"

            # Increment the open style counter
            open_fgs += 1

          # If we have a previous node, and the last node's fg style doesn't
          # match this nodes, then we start a new escape character.
          elsif previous && previous.fg != node.fg
            # If this fg is different to the last fg, and this fg is nil, that means
            # the styling has stopped.
            if !node.fg
              # Add our reset escape character
              buffer << "\e[0m"

              # Decrement the open style counter
              open_fgs -= 1
            else
              buffer << "\e[#{node.fg}m"

            # Increment the open style counter
              open_fgs += 1
            end
          end

          # Add the nodes blob to te buffer
          buffer << node.blob

          # Set this node as being the previous node
          previous = node
        end

        # Be sure to close off any open fg's for this line
        open_fgs.times { buffer << "\e[0m" }

        # Add a new line as long as this line isn't the last
        buffer << "\n" if line_index != last_line_index
      end

      buffer.join("")
    end
  end
end
