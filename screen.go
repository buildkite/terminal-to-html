package terminal

import (
	"math"
	"strconv"
	"strings"
	"sync"
)

// A terminal 'screen'. Current cursor position, cursor style, and characters
type screen struct {
	x          int
	y          int
	screen     []screenLine
	style      *style
	dirtyMutex sync.Mutex
}

type screenLine struct {
	dirty bool
	nodes []node
}
type dirtyLine struct {
	Y    int    `json:"y"`
	HTML string `json:"html"`
}

const screenEndOfLine = -1
const screenStartOfLine = 0

// Clear part (or all) of a line on the screen
func (s *screen) clear(y int, xStart int, xEnd int) {
	if len(s.screen) <= y {
		return
	}
	s.screen[y].dirty = true

	if xStart == screenStartOfLine && xEnd == screenEndOfLine {
		s.screen[y].nodes = make([]node, 0, 80)
	} else {
		line := s.screen[y]

		if xEnd == screenEndOfLine {
			xEnd = len(line.nodes) - 1
		}
		for i := xStart; i <= xEnd && i < len(line.nodes); i++ {
			line.nodes[i] = emptyNode
		}
	}
}

func (s *screen) flushDirty() []dirtyLine {
	lines := make([]dirtyLine, 0, 0)

	s.dirtyMutex.Lock()
	for y := range s.screen {
		if s.screen[y].dirty {
			s.screen[y].dirty = false
			lines = append(lines, dirtyLine{Y: y, HTML: outputLineAsHTML(s.screen[y].nodes)})
		}
	}
	s.dirtyMutex.Unlock()

	return lines
}

func (s *screen) flushAll() []dirtyLine {
	lines := make([]dirtyLine, len(s.screen))

	s.dirtyMutex.Lock()
	for y := range s.screen {
		lines[y] = dirtyLine{Y: y, HTML: outputLineAsHTML(s.screen[y].nodes)}
	}
	s.dirtyMutex.Unlock()

	return lines
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

// Move the cursor down
func (s *screen) down(i string) {
	s.y += pi(i)
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
		s.screen = append(s.screen, screenLine{dirty: true})
		s.screen[i].nodes = make([]node, 0, 80)
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

	line := s.screen[s.y].nodes
	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style}
	s.screen[s.y].dirty = true
	s.screen[s.y].nodes = line
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

func (s *screen) appendElement(i *element) {
	s.growScreenHeight()
	line := s.growLineWidth(s.screen[s.y].nodes)

	line[s.x] = node{style: s.style, elem: i}
	s.screen[s.y].dirty = true
	s.screen[s.y].nodes = line
	s.x++
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
		lines = append(lines, outputLineAsHTML(line.nodes))
	}

	return []byte(strings.Join(lines, "\n"))
}

func (s *screen) newLine() {
	s.x = 0
	s.y++
}

func (s *screen) carriageReturn() {
	s.x = 0
}

func (s *screen) backspace() {
	if s.x > 0 {
		s.x--
	}
}
