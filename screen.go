package terminal

import (
	"math"
	"strconv"
	"strings"
)

// A terminal 'screen'. Current cursor position, cursor style, and characters
type Screen struct {
	// Current cursor position
	x, y int

	// Screen contents
	screen []screenLine

	// Current style
	style style

	// Parser to use for streaming processing
	parser *parser

	// Optional maximum amount of backscroll to retain in the buffer.
	// Setting to 0 or negative makes the screen buffer unlimited.
	MaxLines int

	// Optional callback. If not nil, as each line is scrolled out of the top of
	// the buffer, this func is called with the HTML.
	ScrollOutFunc func(lineHTML string)

	// Processing statistics
	LinesScrolledOut int // count of lines that scrolled off the top
	CursorUpOOB      int // count of times ESC [A or ESC [F tried to move y < 0
	CursorBackOOB    int // count of times ESC [D tried to move x < 0
}

type screenLine struct {
	nodes []node

	// metadata is { namespace => { key => value, ... }, ... }
	// e.g. { "bk" => { "t" => "1234" } }
	metadata map[string]map[string]string

	// element nodes refer to elements in this slice by index
	// (if node.style.element(), then elements[node.blob] is the element)
	elements []*element
}

const (
	screenStartOfLine = 0
	screenEndOfLine   = math.MaxInt
)

// Clear part (or all) of a line on the screen. The range to clear is inclusive
// of xStart and xEnd.
func (s *Screen) clear(y, xStart, xEnd int) {
	if y < 0 || y >= len(s.screen) {
		return
	}

	if xStart < 0 {
		xStart = 0
	}
	if xEnd < xStart {
		// Not a valid range.
		return
	}

	line := &s.screen[y]

	if xStart >= len(line.nodes) {
		// Clearing part of a line starting after the end of the current line...
		return
	}

	if xEnd >= len(line.nodes)-1 {
		// Clear from start to end of the line
		line.nodes = line.nodes[:xStart]
		return
	}

	for i := xStart; i <= xEnd; i++ {
		line.nodes[i] = emptyNode
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
func (s *Screen) up(i string) {
	s.y -= ansiInt(i)
	if s.y < 0 {
		s.CursorUpOOB++
		s.y = 0
	}
}

// Move the cursor down
func (s *Screen) down(i string) {
	s.y += ansiInt(i)
}

// Move the cursor forward on the line
func (s *Screen) forward(i string) {
	s.x += ansiInt(i)
}

// Move the cursor backward, if we can
func (s *Screen) backward(i string) {
	s.x -= ansiInt(i)
	if s.x < 0 {
		s.CursorBackOOB++
		s.x = 0
	}
}

func (s *Screen) getCurrentLineForWriting() *screenLine {
	// Ensure there are enough lines on screen for the cursor's Y position.
	for s.y >= len(s.screen) {
		// If MaxLines is not in use, or adding a new line would not make it
		// larger than MaxLines, then just allocate a new line.
		if s.MaxLines <= 0 || len(s.screen)+1 <= s.MaxLines {
			// nodes is preallocated with space for 80 columns, which is
			// arbitrary, but also the traditional terminal width.
			newLine := screenLine{nodes: make([]node, 0, 80)}
			s.screen = append(s.screen, newLine)
			continue
		}

		// MaxLines is in effect, and adding a new line would make the screen
		// larger than MaxLines.
		// Pass the line being scrolled out to ScrollOutFunc if it exists.
		if s.ScrollOutFunc != nil {
			s.ScrollOutFunc(s.screen[0].asHTML())
		}
		s.LinesScrolledOut++

		// Trim the first line off the top of the screen.
		// Recycle its nodes slice to make a new line on the bottom.
		newLine := screenLine{nodes: s.screen[0].nodes[:0]}
		s.screen = append(s.screen[1:], newLine)
		s.y--
	}

	line := &s.screen[s.y]

	// Add columns if currently shorter than the cursor's x position
	for i := len(line.nodes); i <= s.x; i++ {
		line.nodes = append(line.nodes, emptyNode)
	}
	return line
}

// Write a character to the screen's current X&Y, along with the current screen style
func (s *Screen) write(data rune) {
	line := s.getCurrentLineForWriting()
	line.nodes[s.x] = node{blob: data, style: s.style}
}

// Append a character to the screen
func (s *Screen) append(data rune) {
	s.write(data)
	s.x++
}

// Append multiple characters to the screen
func (s *Screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

func (s *Screen) appendElement(i *element) {
	line := s.getCurrentLineForWriting()
	idx := len(line.elements)
	line.elements = append(line.elements, i)
	ns := s.style
	ns.setElement(true)
	line.nodes[s.x] = node{blob: rune(idx), style: ns}
	s.x++
}

// Set line metadata. Merges the provided data into any existing
// metadata for the current line, overwriting data when keys collide.
func (s *Screen) setLineMetadata(namespace string, data map[string]string) {
	line := s.getCurrentLineForWriting()
	if line.metadata == nil {
		line.metadata = map[string]map[string]string{
			namespace: data,
		}
		return
	}

	ns := line.metadata[namespace]
	if ns == nil {
		// namespace did not exist, set all data
		line.metadata[namespace] = data
		return
	}

	// copy new data over old data
	for k, v := range data {
		ns[k] = v
	}
}

// Apply color instruction codes to the screen's current style
func (s *Screen) color(i []string) {
	s.style = s.style.color(i)
}

// Apply an escape sequence to the screen
func (s *Screen) applyEscape(code rune, instructions []string) {
	if len(instructions) == 0 {
		// Ensure we always have a first instruction
		instructions = []string{""}
	}

	switch code {
	case 'A': // Cursor Up: go up n
		s.up(instructions[0])

	case 'B': // Cursor Down: go down n
		s.down(instructions[0])

	case 'C': // Cursor Forward: go right n
		s.forward(instructions[0])

	case 'D': // Cursor Back: go left n
		s.backward(instructions[0])

	case 'E': // Cursor Next Line: Go to beginning of line n down
		s.x = 0
		s.down(instructions[0])

	case 'F': // Cursor Previous Line: Go to beginning of line n up
		s.x = 0
		s.up(instructions[0])

	case 'G': // Cursor Horizontal Absolute: Go to column n (default 1)
		s.x = max(0, ansiInt(instructions[0])-1)

	// NOTE: H (Cursor Position) is not yet implemented

	case 'J': // Erase in Display: Clears part of the screen.
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

	case 'K': // Erase in Line: erases part of the line.
		switch instructions[0] {
		case "0", "":
			s.clear(s.y, s.x, screenEndOfLine)
		case "1":
			s.clear(s.y, screenStartOfLine, s.x)
		case "2":
			s.clear(s.y, screenStartOfLine, screenEndOfLine)
		}

	case 'M':
		s.color(instructions)
	}
}

// Write writes ANSI text to the screen.
func (s *Screen) Write(input []byte) (int, error) {
	if s.parser == nil {
		s.parser = &parser{
			mode:   parserModeNormal,
			screen: s,
		}
	}
	s.parser.parseToScreen(input)
	return len(input), nil
}

// AsHTML returns the contents of the current screen buffer as HTML.
func (s *Screen) AsHTML() string {
	lines := make([]string, 0, len(s.screen))

	for _, line := range s.screen {
		lines = append(lines, line.asHTML())
	}

	return strings.Join(lines, "\n")
}

// AsPlainText renders the screen without any ANSI style etc.
func (s *Screen) AsPlainText() string {
	lines := make([]string, 0, len(s.screen))

	for _, line := range s.screen {
		lines = append(lines, line.asPlain())
	}

	return strings.Join(lines, "\n")
}

func (s *Screen) newLine() {
	s.x = 0
	s.y++
}

func (s *Screen) revNewLine() {
	if s.y > 0 {
		s.y--
	}
}

func (s *Screen) carriageReturn() {
	s.x = 0
}

func (s *Screen) backspace() {
	if s.x > 0 {
		s.x--
	}
}
