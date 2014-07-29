# A fake terminal screen. Usage is like this:
#
# screen = Screen.new
# screen.x = 10
# screen.y = 10
# screen.write('h')
#
# Will essentially create a 10 x 10 grid with empty characters, and at the
# 10,10 spot element, there will be a 'h'. Co-ordinates start at 0,0
#
# It also supports writing colors. So if you were to change the color like so:
#
# screen.color("42")
#
# Ever new character that you write, will be stored with that color
# information.
#
# When turned into a string, the screen class creates ANSI escape characters,
# that the renderer class gsubs out.
#
# \e[fg32;bg42;
# \e[fgi91;;
# \e[fgx102;bgx102;
# \e[0m
#
# Are some of the examples of the escape sequences that this will render.

module Terminal
  class Screen
    class Node < Struct.new(:blob, :fg, :bg)
      def ==(value)
        blob == value
      end

      # Every node has a style, a foreground style, and a background
      # style. This method returns what essentially becomes the escape
      # sequence:
      #
      # \e[fg;bg;
      #
      # As the screen is turned into a string, the style is used to compare
      # whether or not a new escape sequence is required.
      def style
        if fg || bg
          "#{fg};#{bg};"
        else
          nil
        end
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

      # If the line isn't there, we can't clean it.
      return if line.nil?

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
    def color(color)
      # Reset all styles
      if color == "0"
        @fg = nil
        @bg = nil
        return color
      end

      # Reset foreground color only
      if color == "39"
        @fg = nil
        return color
      end

      # Reset background color only
      if color == "49"
        @bg = nil
        return color
      end

      colors = color.to_s.split(";")

      # Extended set foreground x-term color
      if colors[0] == "38" && colors[1] == "5"
        return @fg = "fgx#{colors[2]}"
      end

      # Extended set background x-term color
      if colors[0] == "48" && colors[1] == "5"
        return @bg = "bgx#{colors[2]}"
      end

      # If multiple colors are defined, i.e. \e[30;42m\e
      # then loop through each one, and assign it to @fg
      # or @bg
      colors.each do |c|
        # If the number is between 30–37, then it's a foreground color,
        # if it's 40–47, then it's a background color. 90-97 is like the regular
        # foreground 30-37, but it's high intensity
        case c.to_i
        when 30..37
          @fg = "fg#{c}"
        when 40..47
          @bg = "bg#{c}"
        when 90..97
          @fg = "fgi#{c}"
        end
      end
    end

    def up(value = nil)
      self.y -= parse_integer(value)
    end

    def down(value = nil)
      new_y = @y + parse_integer(value)

      # Only jump down if the line exists
      if @screen[new_y]
        self.y = new_y
      else
        false
      end
    end

    def backward(value = nil)
      self.x -= parse_integer(value)
    end

    def foward(value = nil)
      self.x += parse_integer(value)
    end

    def to_a
      @screen.to_a.map { |chars| chars.map(&:to_s) }
    end

    # Renders each node to a string. This looks at each node, and then inserts
    # escape characters that will be gsubed into <span> elements.
    #
    # ANSI codes generally span across lines. So if you \e[12m\n\nhello, the hello will
    # inhert the styles of \e[12m. This doesn't work so great in HTML, especially if you
    # wrap divs around each line, so this method also copies any styles that are left open
    # at the end of a line, to the begining of new lines, so you end up with something like this:
    #
    # \e[12m\n\e[12m\n\e[12mhello
    #
    # It also attempts to only insert escapes that are required. Given the following:
    #
    # \e[12mh\e[12me\e[12ml\e[12ml\e[12mo\e[0m
    #
    # A general purpose ANSI renderer will convert it to:
    #
    # <span class="c12">h<span class="c12">e<span class="c12">l<span class="c12">l<span class="c12">o</span></span></span></span>
    #
    # But ours is smart, and tries to do stuff like this:
    #
    # <span class="c12">hello</span>
    def to_s
      last_line_index = @screen.length - 1
      buffer = []

      @screen.each_with_index do |line, line_index|
        previous = nil

        # Keep track of every open style we have, so we know
        # that we need to close any open ones at the end.
        open_styles = 0

        line.each do |node|
          # If there is no previous node, and the current node has a color
          # (first node in a line) then add the escape character.
          if !previous && node.style
            buffer << "\e[#{node.style}m"

            # Increment the open style counter
            open_styles += 1

          # If we have a previous node, and the last node's style doesn't
          # match this nodes, then we start a new escape character.
          elsif previous && previous.style != node.style
            # If the new node has no style, that means that all the styles
            # have been closed.
            if !node.style
              # Add our reset escape character
              buffer << "\e[0m"

              # Decrement the open style counter
              open_styles -= 1
            else
              buffer << "\e[#{node.style}m"

              # Increment the open style counter
              open_styles += 1
            end
          end

          # Add the nodes blob to te buffer
          buffer << node.blob

          # Set this node as being the previous node
          previous = node
        end

        # Be sure to close off any open styles for this line
        open_styles.times { buffer << "\e[0m" }

        # Add a new line as long as this line isn't the last
        buffer << "\n" if line_index != last_line_index
      end

      buffer.join("")
    end

    private

    def parse_integer(value)
      value.nil? || value == "" ? 1 : value.to_i
    end
  end
end
