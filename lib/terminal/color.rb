module Terminal
  class Color < Terminal::Node
    def to_debug
      "[color:#{@code}]"
    end
  end
end
