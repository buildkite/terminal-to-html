class TerminalToHtml < FPM::Cookery::Recipe
  homepage 'https://buildkite.github.io/terminal'

  name     'terminal-to-html'
  version  ENV['RELEASE_VERSION']
  description 'Converts arbitrary shell output (with ANSI) into beautifully rendered HTML'

  def install
    bin.install ['terminal-to-html']
  end
end