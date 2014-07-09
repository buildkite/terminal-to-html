class Fixture
  def self.path
    File.join(File.expand_path(File.dirname(__FILE__)), '..', 'fixtures')
  end

  def self.for(name)
    raw = File.read(File.join(path, "#{name}.raw"))
    rendered = File.read(File.join(path, "#{name}.rendered"))

    new(raw, rendered)
  end

  attr_reader :raw, :rendered

  def initialize(raw, rendered)
    @raw = raw
    @rendered = rendered
  end
end
