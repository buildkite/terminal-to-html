# encoding: UTF-8

module Terminal
  module Cache
    extend self

    def cache(table, key, &block)
      @cache ||= {}
      @cache[table] ||= {}
      @cache[table][key] ||= yield
    end
  end
end
