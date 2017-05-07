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
	"encoding/json"
	"log"
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

func (s *Streamer) Flush(all bool) [][]byte {
	s.mutex.Lock()
	var dirtyLines []dirtyLine
	if all {
		dirtyLines = s.screen.flushAll()
	} else {
		dirtyLines = s.screen.flushDirty()
	}
	s.mutex.Unlock()

	output := make([][]byte, len(dirtyLines))
	for idx, line := range dirtyLines {
		lineOut, err := json.Marshal(line)
		if err != nil {
			log.Fatalf("Couldn't encode %q to JSON: %s", line, err)
		}
		output[idx] = lineOut
	}
	return output
}
