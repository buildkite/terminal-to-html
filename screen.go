package terminal

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// A terminal 'screen'. Current cursor position, cursor style, and characters
type screen struct {
	x      int
	y      int
	screen [][]node
	style  *style
}

// Stateful container object for capturing escape codes
type escapeCode struct {
	instructions    []string
	buffer          []rune
	nextInstruction []rune
	code            rune
}

const screenEndOfLine = -1
const screenStartOfLine = 0

// Clear part (or all) of a line on the screen
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

// "Safe" parseint for parsing ANSI instructions
func pi(s string) int {
	if s == "" {
		return 1
	}
	i, _ := strconv.ParseInt(s, 10, 8)
	return int(i)
}

// Move the cursor up, if we can
func (s *screen) up(i string) {
	s.y -= pi(i)
	s.y = int(math.Max(0, float64(s.y)))
}

// Move the cursor down, if we can
func (s *screen) down(i string) {
	s.y += pi(i)
	s.y = int(math.Min(float64(s.y), float64(len(s.screen)-1)))
}

// Move the cursor forward on the line
func (s *screen) forward(i string) {
	s.x += pi(i)
}

// Move the cursor backward, if we can
func (s *screen) backward(i string) {
	s.x -= pi(i)
	s.x = int(math.Max(0, float64(s.x)))
}

// Add rows to our screen if necessary
func (s *screen) growScreenHeight() {
	for i := len(s.screen); i <= s.y; i++ {
		s.screen = append(s.screen, make([]node, 0, 80))
	}
}

// Add columns to our current line if necessary
func (s *screen) growLineWidth(line []node) []node {
	for i := len(line); i <= s.x; i++ {
		line = append(line, emptyNode)
	}
	return line
}

// Write a character to the screen's current X&Y, along with the current screen style
func (s *screen) write(data rune) {
	s.growScreenHeight()

	line := s.screen[s.y]
	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style}
	s.screen[s.y] = line
}

// Append a character to the screen
func (s *screen) append(data rune) {
	s.write(data)
	s.x++
}

// Append multiple characters to the screen
func (s *screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

// Apply color instruction codes to the screen's current style
func (s *screen) color(i []string) {
	s.style = s.style.color(i)
}

// Apply an escape sequence to the screen
func (s *screen) applyEscape(e escapeCode) {
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

// Reset our instruction buffer & add to our instruction list
func (e *escapeCode) endOfInstruction() {
	e.instructions = append(e.instructions, string(e.nextInstruction))
	e.nextInstruction = []rune{}
}

// First instruction for this escape code, if we have one.
func (e *escapeCode) firstInstruction() string {
	if len(e.instructions) == 0 {
		return ""
	}
	return e.instructions[0]
}

// Accept ANSI input and turn in to a series of nodes on our screen.
func (s *screen) render(input []byte) {
	s.style = &emptyStyle
	insideEscapeCode := false
	var escape escapeCode

	// TODO: Ugh.
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
					s.applyEscape(escape)
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

func (s *screen) output() []byte {
	var lines []string

	for _, line := range s.screen {
		var openStyles int
		var lineBuf outputBuffer

		for idx, node := range line {
			if idx == 0 && !node.style.isEmpty() {
				lineBuf.appendNodeStyle(node)
				openStyles++
			} else if idx > 0 {
				previous := line[idx-1]
				if !node.hasSameStyle(previous) {
					if node.style.isEmpty() {
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
