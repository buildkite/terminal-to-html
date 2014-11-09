package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var emptyLineRegex = regexp.MustCompile(`^$`)

const screenEndOfLine = -1
const screenStartOfLine = 0

var emptyStyle = style{}
var emptyNode = node{' ', &emptyStyle}

type node struct {
	blob  uint8
	style *style
}

type screen struct {
	x      int
	y      int
	screen [][]node
	style  *style
}

type style struct {
	fgColor     string
	bgColor     string
	otherColors []string
	asString    string
}

func (n *node) hasSameStyle(o node) bool {
	return n.style.String() == o.style.String()
}

func appendNodeStyle(b *bytes.Buffer, n node) {
	b.Write([]byte(`<span class="`))
	b.Write([]byte(n.style.String()))
	b.Write([]byte(`">`))
}

func closeStyle(b *bytes.Buffer) {
	b.Write([]byte("</span>"))
}

func (s *style) emptyNode() bool {
	return s.fgColor == "" && s.bgColor == "" && len(s.otherColors) == 0
}

func (s *screen) output() []byte {
	var lines []string

	for _, line := range s.screen {
		var openStyles int
		var lineBuf bytes.Buffer

		for idx, node := range line {
			if idx == 0 && !node.style.emptyNode() {
				appendNodeStyle(&lineBuf, node)
				openStyles++
			} else if idx > 0 {
				previous := line[idx-1]
				if !node.hasSameStyle(previous) {
					if node.style.emptyNode() {
						closeStyle(&lineBuf)
						openStyles--
					} else {
						appendNodeStyle(&lineBuf, node)
						openStyles++
					}
				}
			}
			appendChar(&lineBuf, node.blob)
		}
		for i := 0; i < openStyles; i++ {
			closeStyle(&lineBuf)
		}
		asString := strings.TrimRight(lineBuf.String(), " \t")

		lines = append(lines, asString)
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

func appendChar(b *bytes.Buffer, char byte) {
	switch char {
	case '&':
		b.WriteString("&amp;")
	case '\'':
		b.WriteString("&#39;")
	case '<':
		b.WriteString("&lt;")
	case '>':
		b.WriteString("&gt;")
	case '"':
		b.WriteString("&#34;")
	default:
		b.WriteByte(char)
	}
}

func remove(a []string, r string) []string {
	// Must be a better way ..
	var removed []string

	for _, s := range a {
		if s != r {
			removed = append(removed, s)
		}
	}
	return removed
}

func (screen *screen) color(i string) {
	colors := strings.Split(i, ";")

	s := new(style)
	if screen.style != nil {
		s.fgColor = screen.style.fgColor
		s.bgColor = screen.style.bgColor
		s.otherColors = screen.style.otherColors
	}
	screen.style = s

	if len(colors) >= 2 {
		if colors[0] == "38" && colors[1] == "5" {
			// Extended set foreground x-term color
			s.fgColor = "term-fgx" + colors[2]
			return
		}

		// Extended set background x-term color
		if colors[0] == "48" && colors[1] == "5" {
			s.bgColor = "term-bgx" + colors[2]
			return
		}
	}

	for _, cc := range colors {
		// // If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// // one, and assign it to s.fgColor or s.bgColor
		cInteger := pi(cc)
		if cInteger == 0 {
			// Reset all styles
			s.fgColor = ""
			s.bgColor = ""
			s.otherColors = make([]string, 0)
			// Primary (default) font
		} else if cInteger == 10 {
			// no-op
			// Turn off bold / Normal color or intensity (21 & 22 essentially do the same thing)
		} else if cInteger == 21 || cInteger == 22 {
			s.otherColors = remove(s.otherColors, "term-fg1")
			s.otherColors = remove(s.otherColors, "term-fg2")
			// Turn off italic
		} else if cInteger == 23 {
			s.otherColors = remove(s.otherColors, "term-fg3")
			// Turn off underline
		} else if cInteger == 24 {
			s.otherColors = remove(s.otherColors, "term-fg4")
			// Turn off crossed-out
		} else if cInteger == 29 {
			s.otherColors = remove(s.otherColors, "term-fg9")
			// Reset foreground color only
		} else if cInteger == 39 {
			s.fgColor = ""
			// Reset background color only
		} else if cInteger == 49 {
			s.bgColor = ""
			// 30–37, then it's a foreground color
		} else if cInteger >= 30 && cInteger <= 37 {
			s.fgColor = "term-fg" + cc
			// 40–47, then it's a background color.
		} else if cInteger >= 40 && cInteger <= 47 {
			s.bgColor = "term-bg" + cc
			// 90-97 is like the regular fg color, but high intensity
		} else if cInteger >= 90 && cInteger <= 97 {
			s.fgColor = "term-fgi" + cc
			// 100-107 is like the regular bg color, but high intensity
		} else if cInteger >= 100 && cInteger <= 107 {
			s.fgColor = "term-bgi" + cc
			// 1-9 random other styles
		} else if cInteger >= 1 && cInteger <= 9 {
			s.otherColors = append(s.otherColors, "term-fg"+cc)
		}
	}
}

func (s *screen) clear(y int, xStart int, xEnd int) {
	if len(s.screen) < y {
		return
	}

	if xStart == screenStartOfLine && xEnd == screenEndOfLine {
		s.screen[y] = make([]node, 0, 80)
	} else {
		line := s.screen[y]

		if xEnd == screenEndOfLine {
			xEnd = len(line) - 1
		}
		for i := xStart; i <= xEnd; i++ {
			line[i] = emptyNode
		}
	}
}

func pi(s string) int {
	i, err := strconv.ParseInt(s, 10, 8)
	check(err)
	return int(i)
}

func (s *screen) up(i string) {
	s.y -= pi(i)
	s.y = int(math.Max(0, float64(s.y)))
}

func (s *screen) down(i string) {
	s.y += pi(i)
	s.y = int(math.Min(float64(s.y), float64(len(s.screen))))
}

func (s *screen) forward(i string) {
	s.x += pi(i)
}

func (s *screen) backward(i string) {
	s.x -= pi(i)
	s.x = int(math.Max(0, float64(s.x)))
}

func (s *screen) growScreenHeight() {
	for i := len(s.screen); i <= s.y; i++ {
		s.screen = append(s.screen, make([]node, 0, 80))
	}
}

func (s *screen) growLineWidth(line []node) []node {
	for i := len(line); i <= s.x; i++ {
		line = append(line, emptyNode)
	}
	return line
}

func (s *style) String() string {
	if s.asString != "" || s.emptyNode() {
		return s.asString
	}

	var styles []string
	if s.fgColor != "" {
		styles = append(styles, s.fgColor)
	}
	if s.bgColor != "" {
		styles = append(styles, s.bgColor)
	}
	styles = append(styles, s.otherColors...)
	s.asString = strings.Join(styles, " ")
	return s.asString
}

func (s *screen) write(data uint8) {
	s.growScreenHeight()

	line := s.screen[s.y]
	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style}
	s.screen[s.y] = line
}

func (s *screen) append(data uint8) {
	s.write(data)
	s.x++
}

func convertToHTML(input string) string {
	return emptyLineRegex.ReplaceAllLiteralString(input, "&nbsp;")
}

func renderToScreen(input []byte) string {
	var screen screen
	screen.style = &emptyStyle
	for i := 0; i < len(input); i++ {
		char := input[i]
		if char == '\n' {
			screen.x = 0
			screen.y++
		} else if char == '\r' {
			screen.x = 0
		} else if char == '\b' {
			screen.x--
		} else if char == '\x1b' {
			len, instruction, code := captureEscapeCode(input[i+1 : i+50])
			i += len

			if code == ' ' {
				// noop
			} else if code == 'm' {
				screen.color(instruction)
			} else if code == 'G' || code == 'g' {
				screen.x = 0
			} else if code == 'K' || code == 'k' {
				if instruction == "" || instruction == "0" {
					screen.clear(screen.y, screen.x, screenEndOfLine)
				} else if instruction == "1" {
					screen.clear(screen.y, screenStartOfLine, screen.x)
				} else if instruction == "2" {
					screen.clear(screen.y, screenStartOfLine, screenEndOfLine)
				}
			} else if code == 'A' {
				screen.up(instruction)
			} else if code == 'B' {
				screen.down(instruction)
			} else if code == 'C' {
				screen.forward(instruction)
			} else if code == 'D' {
				screen.backward(instruction)
			}
		} else {
			screen.append(char)
		}
	}
	return string(screen.output())
}

func captureEscapeCode(input []byte) (length int, instruction string, code byte) {
	codeIndex := bytes.IndexAny(input, "qQmKGgKAaBbCcDd")
	if codeIndex == -1 {
		return 0, "", ' '
	}
	return codeIndex + 1, string(input[1:codeIndex]), input[codeIndex]
}

func Render(input []byte) string {
	output := renderToScreen(input)
	output = convertToHTML(output)
	return output
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-serve" {
			http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
				input, err := ioutil.ReadAll(r.Body)
				check(err)
				// t0 := time.Now()
				output := Render(input)
				// t1 := time.Now()
				w.Write([]byte(output))
				// log.Printf("recv %d bytes, xmit %d bytes, took %v", len(input), len(output), t1.Sub(t0))
			})

			log.Printf("Listening on port 1337")
			log.Fatal(http.ListenAndServe(":1337", nil))
		} else {
			input, err := ioutil.ReadFile(os.Args[1])
			check(err)
			fmt.Printf("%v", Render(input))
		}
	} else {
		input, err := ioutil.ReadAll(os.Stdin)
		check(err)
		fmt.Printf("%v", Render(input))
	}
}
