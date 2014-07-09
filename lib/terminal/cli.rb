module Terminal
  class CLI
    def self.run(*args, file)
      new(args, file).run
    end

    def initialize(args, file)
      @args = args
      @file = file
    end

    def run
      puts @file
    end
  end
end
