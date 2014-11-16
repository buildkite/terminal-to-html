/*
Package terminal converts ANSI input to HTML output.

The generated HTML needs to be used with the stylesheet at
https://raw.githubusercontent.com/buildbox/terminal/master/app/assets/stylesheets/terminal.css
and wrapped in a term-container div.
*/
package terminal

import "bytes"

type screen struct {
	x      int
	y      int
	screen [][]node
	style  *style
}

// Render converts ANSI to HTML and returns the result.
func Render(input []byte) []byte {
	screen := screen{}
	screen.render(input)
	output := bytes.Replace(screen.output(), []byte("\n\n"), []byte("\n&nbsp;\n"), -1)
	return output
}
