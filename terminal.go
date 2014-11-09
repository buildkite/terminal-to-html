package terminal

import (
	"bytes"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var emptyLineRegex = regexp.MustCompile(`^$`)

const screenEndOfLine = -1
const screenStartOfLine = 0

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

type outputBuffer struct {
	buf bytes.Buffer
}

func (n *node) hasSameStyle(o node) bool {
	return n.style.string() == o.style.string()
}

func (b *outputBuffer) appendNodeStyle(n node) {
	b.buf.Write([]byte(`<span class="`))
	b.buf.Write([]byte(n.style.string()))
	b.buf.Write([]byte(`">`))
}

func (b *outputBuffer) closeStyle() {
	b.buf.Write([]byte("</span>"))
}

func (s *screen) output() []byte {
	var lines []string

	for _, line := range s.screen {
		var openStyles int
		var lineBuf outputBuffer

		for idx, node := range line {
			if idx == 0 && !node.style.empty() {
				lineBuf.appendNodeStyle(node)
				openStyles++
			} else if idx > 0 {
				previous := line[idx-1]
				if !node.hasSameStyle(previous) {
					if node.style.empty() {
						lineBuf.closeStyle()
						openStyles--
					} else {
						lineBuf.appendNodeStyle(node)
						openStyles++
					}
				}
			}
			lineBuf.appendChar(node.blob)
		}
		for i := 0; i < openStyles; i++ {
			lineBuf.closeStyle()
		}
		asString := strings.TrimRight(lineBuf.buf.String(), " \t")

		lines = append(lines, asString)
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

func (b *outputBuffer) appendChar(char byte) {
	switch char {
	case '&':
		b.buf.WriteString("&amp;")
	case '\'':
		b.buf.WriteString("&#39;")
	case '<':
		b.buf.WriteString("&lt;")
	case '>':
		b.buf.WriteString("&gt;")
	case '"':
		b.buf.WriteString("&#34;")
	default:
		b.buf.WriteByte(char)
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

func (s *screen) color(i string) {
	s.style = s.style.color(i)
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
