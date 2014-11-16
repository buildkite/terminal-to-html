package terminal

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const screenEndOfLine = -1
const screenStartOfLine = 0

var emptyNode = node{' ', &emptyStyle}

type node struct {
	blob  rune
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
	return n == &o || n.style.string() == o.style.string()
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

	return []byte(strings.Join(lines, "\n"))
}

func (b *outputBuffer) appendChar(char rune) {
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
		b.buf.WriteString("&quot;")
	case '/':
		b.buf.WriteString("&#47;")
	default:
		b.buf.WriteRune(char)
	}
}

func (s *screen) clear(y int, xStart int, xEnd int) {
	if len(s.screen) <= y {
		return
	}

	if xStart == screenStartOfLine && xEnd == screenEndOfLine {
		s.screen[y] = make([]node, 0, 80)
	} else {
		line := s.screen[y]

		if xEnd == screenEndOfLine {
			xEnd = len(line) - 1
		}
		for i := xStart; i <= xEnd && i < len(line); i++ {
			line[i] = emptyNode
		}
	}
}

func pi(s string) int {
	if s == "" {
		return 1
	}
	i, _ := strconv.ParseInt(s, 10, 8)
	return int(i)
}

func (s *screen) up(i string) {
	s.y -= pi(i)
	s.y = int(math.Max(0, float64(s.y)))
}

func (s *screen) down(i string) {
	s.y += pi(i)
	s.y = int(math.Min(float64(s.y), float64(len(s.screen)-1)))
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

func (s *screen) write(data rune) {
	s.growScreenHeight()

	line := s.screen[s.y]
	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style}
	s.screen[s.y] = line
}

func (s *screen) append(data rune) {
	s.write(data)
	s.x++
}

func (s *screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

func (s *screen) color(i []string) {
	s.style = s.style.color(i)
}

type escapeCode struct {
	instructions    []string
	buffer          []rune
	nextInstruction []rune
	code            rune
}

func (e *escapeCode) endOfInstruction() {
	e.instructions = append(e.instructions, string(e.nextInstruction))
	e.nextInstruction = []rune{}
}

func (e *escapeCode) firstInstruction() string {
	if len(e.instructions) == 0 {
		return ""
	}
	return e.instructions[0]
}

func (s *screen) render(input []byte) {
	s.style = &emptyStyle
	insideEscapeCode := false
	var escape escapeCode
	for _, char := range string(input) {
		if insideEscapeCode {
			escape.buffer = append(escape.buffer, char)
			if len(escape.buffer) == 2 {
				if char != '[' {
					// Not really an escape code, abort
					s.appendMany(escape.buffer)
					insideEscapeCode = false
				}
			} else {
				char = unicode.ToUpper(char)
				switch char {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					escape.nextInstruction = append(escape.nextInstruction, char)
				case ';':
					escape.endOfInstruction()
				case 'Q', 'K', 'G', 'A', 'B', 'C', 'D', 'M':
					escape.code = char
					escape.endOfInstruction()
					escape.applyToScreen(s)
					insideEscapeCode = false
				default:
					// abort the escapeCode
					s.appendMany(escape.buffer)
					insideEscapeCode = false
				}
			}
		} else {
			switch char {
			case '\n':
				s.x = 0
				s.y++
			case '\r':
				s.x = 0
			case '\b':
				if s.x > 0 {
					s.x--
				}
			case '\x1b':
				escape = escapeCode{buffer: []rune{char}}
				insideEscapeCode = true
			default:
				s.append(char)
			}
		}
	}
}

func (e *escapeCode) applyToScreen(s *screen) {
	switch e.code {
	case 'M':
		s.color(e.instructions)
	case 'G':
		s.x = 0
	case 'K':
		switch e.firstInstruction() {
		case "0", "":
			s.clear(s.y, s.x, screenEndOfLine)
		case "1":
			s.clear(s.y, screenStartOfLine, s.x)
		case "2":
			s.clear(s.y, screenStartOfLine, screenEndOfLine)
		}
	case 'A':
		s.up(e.firstInstruction())
	case 'B':
		s.down(e.firstInstruction())
	case 'C':
		s.forward(e.firstInstruction())
	case 'D':
		s.backward(e.firstInstruction())
	}
}

func captureEscapeCode(inputRunes []rune) (length int, instructions string, code byte) {
	maxCaptureLength := int(math.Min(float64(len(inputRunes)), 50))
	input := string(inputRunes[:maxCaptureLength])

	codeIndex := strings.IndexAny(input, " qQmKGgKAaBbCcDd")
	if codeIndex == -1 || input[codeIndex] == ' ' {
		return 0, "", ' '
	}
	return codeIndex + 1, input[1:codeIndex], input[codeIndex]
}

func Render(input []byte) []byte {
	screen := screen{}
	screen.render(input)
	output := bytes.Replace(screen.output(), []byte("\n\n"), []byte("\n&nbsp;\n"), -1)
	return output
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
