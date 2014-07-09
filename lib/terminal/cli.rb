module Terminal
  class CLI
    def self.run(*args)
      new(args).run
    end

    def initialize(args)
      @args = args
    end

    def run
      puts @args
    end
  end
end
