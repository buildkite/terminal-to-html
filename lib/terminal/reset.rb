module Terminal
  class Reset < Terminal::Node
    def initialize(opened)
      @opened = opened
      @code = "0"
    end

    def to_s
      super * @opened
    end

    def to_debug
      "[reset:#{@opened}]"
    end
  end
end
