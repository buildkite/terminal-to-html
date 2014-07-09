module Terminal
  class Node
    def initialize(code)
      @code = code
    end

    def to_s
      "\e[#{@code}m"
    end

    def to_debug
      "[node:#{@code}]"
    end
  end
end
