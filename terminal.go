/*
Package terminal converts ANSI input to HTML output.

The generated HTML needs to be used with the stylesheet at
https://raw.githubusercontent.com/buildkite/terminal-to-html/main/assets/terminal.css
and wrapped in a term-container div.

You can call this library from the command line with terminal-to-html:
GO111MODULE=on go install github.com/buildkite/terminal-to-html/v3/cmd/terminal-to-html
*/
package terminal

import "strings"

// Render converts ANSI to HTML and returns the result.
func Render(input []byte) string {
	screen := Screen{}
	screen.Write(input)
	output := strings.Replace(screen.AsHTML(), "\n\n", "\n&nbsp;\n", -1)
	return output
}
