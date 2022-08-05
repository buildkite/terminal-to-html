package terminal

import (
	"bytes"
	"math"
	"strconv"
	"strings"
)

// A terminal 'screen'. Current cursor position, cursor style, and characters
type screen struct {
	x      int
	y      int
	screen [][]node
	style  *style
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
func ansiInt(s string) int {
	if s == "" {
		return 1
	}
	i, _ := strconv.ParseInt(s, 10, 8)
	return int(i)
}

// Move the cursor up, if we can
func (s *screen) up(i string) {
	s.y -= ansiInt(i)
	s.y = int(math.Max(0, float64(s.y)))
}

// Move the cursor down
func (s *screen) down(i string) {
	s.y += ansiInt(i)
}

// Move the cursor forward on the line
func (s *screen) forward(i string) {
	s.x += ansiInt(i)
}

// Move the cursor backward, if we can
func (s *screen) backward(i string) {
	s.x -= ansiInt(i)
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

	// If the node already exists with a persistent element, retain that element.
	var persistentElement *element
	if s.x <= len(line)-1 && line[s.x].hasPersistentElement() {
		persistentElement = line[s.x].elem
	}

	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style, elem: persistentElement}
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

func (s *screen) appendElement(el *element) {
	s.growScreenHeight()
	line := s.growLineWidth(s.screen[s.y])

	line[s.x] = node{style: s.style, elem: el}
	s.screen[s.y] = line

	if el.elementType == ELEMENT_BK {
		// x is not incremented: we've placed the element node onto the _next_ cell
	} else {
		s.x++
	}
}

// Apply color instruction codes to the screen's current style
func (s *screen) color(i []string) {
	s.style = s.style.color(i)
}

// Apply an escape sequence to the screen
func (s *screen) applyEscape(code rune, instructions []string) {
	if len(instructions) == 0 {
		// Ensure we always have a first instruction
		instructions = []string{""}
	}

	switch code {
	case 'M':
		s.color(instructions)
	case 'G':
		s.x = 0
	// "Erase in Display"
	case 'J':
		switch instructions[0] {
		// "erase from current position to end (inclusive)"
		case "0", "":
			// This line should be equivalent to K0
			s.clear(s.y, s.x, screenEndOfLine)
			// Truncate the screen below the current line
			if len(s.screen) > s.y {
				s.screen = s.screen[:s.y+1]
			}
		// "erase from beginning to current position (inclusive)"
		case "1":
			// This line should be equivalent to K1
			s.clear(s.y, screenStartOfLine, s.x)
			// Truncate the screen above the current line
			if len(s.screen) > s.y {
				s.screen = s.screen[s.y+1:]
			}
			// Adjust the cursor position to compensate
			s.y = 0
		// 2: "erase entire display", 3: "erase whole display including scroll-back buffer"
		// Given we don't have a scrollback of our own, we treat these as equivalent
		case "2", "3":
			s.screen = nil
			s.x = 0
			s.y = 0
		}
	// "Erase in Line"
	case 'K':
		switch instructions[0] {
		case "0", "":
			s.clear(s.y, s.x, screenEndOfLine)
		case "1":
			s.clear(s.y, screenStartOfLine, s.x)
		case "2":
			s.clear(s.y, screenStartOfLine, screenEndOfLine)
		}
	case 'A':
		s.up(instructions[0])
	case 'B':
		s.down(instructions[0])
	case 'C':
		s.forward(instructions[0])
	case 'D':
		s.backward(instructions[0])
	}
}

// Parse ANSI input, populate our screen buffer with nodes
func (s *screen) parse(ansi []byte) {
	s.style = &emptyStyle

	parseANSIToScreen(s, ansi)
}

func (s *screen) asHTML() []byte {
	var lines []string

	for _, line := range s.screen {
		lines = append(lines, outputLineAsHTML(line))
	}

	return []byte(strings.Join(lines, "\n"))
}

// asPlainText renders the screen without any ANSI style etc.
func (s *screen) asPlainText() string {
	var buf bytes.Buffer
	for i, line := range s.screen {
		for _, node := range line {
			if r, ok := node.getRune(); ok {
				buf.WriteRune(r)
			}
		}
		if i < len(s.screen)-1 {
			buf.WriteRune('\n')
		}
	}
	return strings.TrimRight(buf.String(), " \t")
}

func (s *screen) newLine() {
	s.x = 0
	s.y++
}

func (s *screen) revNewLine() {
	if s.y > 0 {
		s.y--
	}
}

func (s *screen) carriageReturn() {
	s.x = 0
}

func (s *screen) backspace() {
	if s.x > 0 {
		s.x--
	}
}
