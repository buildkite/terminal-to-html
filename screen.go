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

const (
	MODE_NORMAL = iota
	MODE_ESCAPE = iota
)

type parser struct {
	mode   int
	escape escapeCode
	screen *screen
}

func newParser(s *screen) parser {
	return parser{
		mode:   MODE_NORMAL,
		screen: s,
	}
}

// Parse ANSI input, populate our screen buffer with nodes
func (s *screen) parse(ansi []byte) {
	s.style = &emptyStyle

	p := newParser(s)

	for _, char := range string(ansi) {
		switch p.mode {
		case MODE_ESCAPE:
			p.parseEscape(char)
		case MODE_NORMAL:
			p.parseNormal(char)
		}
	}
}

func (p *parser) parseEscape(char rune) {
	p.escape.buffer = append(p.escape.buffer, char)
	if len(p.escape.buffer) == 2 {
		if char != '[' {
			// Not really an escape code, abort
			p.screen.appendMany(p.escape.buffer)
			p.mode = MODE_NORMAL
		}
	} else {
		char = unicode.ToUpper(char)
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			p.escape.nextInstruction = append(p.escape.nextInstruction, char)
		case ';':
			p.escape.endOfInstruction()
		case 'Q', 'K', 'G', 'A', 'B', 'C', 'D', 'M':
			p.escape.code = char
			p.escape.endOfInstruction()
			p.screen.applyEscape(p.escape)
			p.mode = MODE_NORMAL
		default:
			// abort the escapeCode
			p.screen.appendMany(p.escape.buffer)
			p.mode = MODE_NORMAL
		}
	}
}

func (p *parser) parseNormal(char rune) {
	switch char {
	case '\n':
		p.screen.x = 0
		p.screen.y++
	case '\r':
		p.screen.x = 0
	case '\b':
		if p.screen.x > 0 {
			p.screen.x--
		}
	case '\x1b':
		p.escape = escapeCode{buffer: []rune{char}}
		p.mode = MODE_ESCAPE
	default:
		p.screen.append(char)
	}
}

func (s *screen) asHTML() []byte {
	var lines []string

	for _, line := range s.screen {
		lines = append(lines, outputLineAsHTML(line))
	}

	return []byte(strings.Join(lines, "\n"))
}
