/*
Package terminal converts ANSI input to HTML output.

The generated HTML needs to be used with the stylesheet at
https://raw.githubusercontent.com/buildkite/terminal/master/assets/terminal.css
and wrapped in a term-container div.

You can call this library from the command line with terminal-to-html:
go install github.com/buildkite/terminal/cmd/terminal-to-html
*/
package terminal

import (
	"bytes"
	"sync"
)

type Streamer struct {
	screen screen
	mutex  sync.Mutex
}

// Render converts ANSI to HTML and returns the result.
func Render(input []byte) []byte {
	screen := screen{}
	screen.parse(input)
	output := bytes.Replace(screen.asHTML(), []byte("\n\n"), []byte("\n&nbsp;\n"), -1)
	return output
}

func (s *Streamer) Write(input []byte) {
	s.mutex.Lock()
	s.screen.parse(input)
	s.mutex.Unlock()
}

func (s *Streamer) Render() []byte {
	return bytes.Replace(s.screen.asHTML(), []byte("\n\n"), []byte("\n&nbsp;\n"), -1)
}

func (s *Streamer) Flush(all bool) []DirtyLine {
	s.mutex.Lock()
	var lines []DirtyLine
	if all {
		lines = s.screen.flushAll()
	} else {
		lines = s.screen.flushDirty()
	}
	s.mutex.Unlock()

	return lines
}
