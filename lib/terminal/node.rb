module Terminal
  class Node
    def initialize(code)
      @code = code
    end

    def to_s
      "\e[#{@code}m"
    end
  end
end
