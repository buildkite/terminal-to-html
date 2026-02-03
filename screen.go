package terminal

import (
	"fmt"
	"maps"
	"math"
	"strconv"
	"strings"
)

const (
	screenStartOfLine = 0
	screenEndOfLine   = math.MaxInt
)

// A terminal 'screen'. Tracks cursor position, cursor style, content, size...
type Screen struct {
	// Current cursor position on the screen
	x, y int

	// Screen contents
	screen []screenLine

	// Current style
	style style

	// Current URL for OSC 8 (iTerm-style) hyperlinking
	urlBrush string

	// Parser to use for streaming processing
	parser parser

	// Optional maximum amount of backscroll to retain in the buffer.
	// Also sets an upper bound on window height.
	// Setting to 0 or negative makes the screen buffer unlimited.
	maxLines int

	// Optional upper bound on window width.
	// Setting to 0 or negative doesn't enforce a limit.
	maxColumns int

	// Current window size. This is required to properly bound cursor movement
	// commands and implement line wrapping.
	// It defaults to 160 columns * 100 lines.
	cols, lines int

	// When multiple screen lines are scrolled out at once, their storage can be
	// recycled later on.
	nodeRecycling [][]node

	// Optional callback. If not nil, as each line is scrolled out of the top of
	// the buffer, this func is called with the HTML.
	// The line will always have a `\n` suffix.
	ScrollOutFunc func(lineHTML string)

	// Processing statistics
	LinesScrolledOut int // count of lines that scrolled off the top
	CursorUpOOB      int // count of times ESC [A or ESC [F tried to move y < 0
	CursorDownOOB    int // count of times ESC [B or ESC [G tried to move y >= height
	CursorFwdOOB     int // count of times ESC [C tried to move x >= width
	CursorBackOOB    int // count of times ESC [D tried to move x < 0
}

// ScreenOption is a functional option for creating new screens.
type ScreenOption = func(*Screen) error

// WithSize sets the initial window size.
func WithSize(w, h int) ScreenOption {
	return func(s *Screen) error { return s.SetSize(w, h) }
}

// WithMaxSize sets the screen size limits.
func WithMaxSize(maxCols, maxLines int) ScreenOption {
	return func(s *Screen) error {
		s.maxColumns, s.maxLines = maxCols, maxLines
		// Ensure the size fits within the new limits.
		if maxCols > 0 {
			s.cols = min(s.cols, maxCols)
		}
		if maxLines > 0 {
			s.lines = min(s.lines, maxLines)
		}
		return nil
	}
}

// NewScreen creates a new screen with various options.
func NewScreen(opts ...ScreenOption) (*Screen, error) {
	s := &Screen{
		// Arbitrarily chosen size, but 160 is double the traditional terminal
		// width (80) and 100 is 4x the traditional terminal height (25).
		// 160x100 also matches the buildkite-agent PTY size.
		cols:  160,
		lines: 100,
		parser: parser{
			mode: parserModeNormal,
		},
	}
	s.parser.screen = s
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// SetSize changes the window size.
func (s *Screen) SetSize(cols, lines int) error {
	if cols <= 0 || lines <= 0 {
		return fmt.Errorf("negative dimension in size %dw x %dh", cols, lines)
	}
	if s.maxColumns > 0 && cols > s.maxColumns {
		return fmt.Errorf("cols greater than max [%d > %d]", cols, s.maxColumns)
	}
	if s.maxLines > 0 && lines > s.maxLines {
		return fmt.Errorf("lines greater than max [%d > %d]", lines, s.maxLines)
	}
	s.cols, s.lines = cols, lines
	return nil
}

// ansiInt parses s as a decimal integer. If s is empty or malformed, it
// returns 1.
func ansiInt(s string) int {
	if s == "" {
		return 1
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return i
}

// Move the cursor up, if we can
func (s *Screen) up(i string) {
	s.y -= ansiInt(i)
	if s.y < 0 {
		s.CursorUpOOB++
		s.y = 0
	}
	// If the cursor was past the end and we change its y position, it moves to
	// the final column instead.
	s.x = min(s.x, s.cols-1)
}

// Move the cursor down, if we can
func (s *Screen) down(i string) {
	s.y += ansiInt(i)
	if s.y >= s.lines {
		s.CursorDownOOB++
		s.y = s.lines - 1
	}
	// If the cursor was past the end and we change its y position, it moves to
	// the final column instead.
	s.x = min(s.x, s.cols-1)
}

// Move the cursor forward (right) on the line, if we can
func (s *Screen) forward(i string) {
	s.x += ansiInt(i)
	if s.x >= s.cols {
		s.CursorFwdOOB++
		s.x = s.cols - 1
	}
}

// Move the cursor backward (left), if we can
func (s *Screen) backward(i string) {
	s.x -= ansiInt(i)
	if s.x < 0 {
		s.CursorBackOOB++
		s.x = 0
	}
}

// top returns the index within s.screen where the window begins.
// The top of the window is not necessarily the top of the buffer: in fact,
// the window is always the bottom-most s.lines (or fewer) elements of s.screen.
// top + s.y = the index of the line where the cursor is.
func (s *Screen) top() int {
	return max(0, len(s.screen)-s.lines)
}

// currentLine returns the line the cursor is on, or nil if no such line has
// been added to the screen buffer yet.
func (s *Screen) currentLine() *screenLine {
	yidx := s.top() + s.y
	if yidx < 0 || yidx >= len(s.screen) {
		return nil
	}
	return &s.screen[yidx]
}

// currentLineForWriting returns the line the cursor is on, or if there is no
// line allocated in the buffer yet, allocates a new line and ensures it has
// enough nodes to write something at the cursor position.
func (s *Screen) currentLineForWriting() *screenLine {
	// If the cursor is past the end, we actually need the line after this one.
	if s.x == s.cols {
		// Handle line wrapping.
		// Doing this at write time allows the cursor to be positioned past the end,
		// as would happen if the entire line (including the last column) was
		// written to, but doesn't allow writing past the last column.
		// Since the cursor cannot be moved to s.cols, this can only happen if we
		// have written all the way to the end of a line, so we can safely assume
		// the current line exists.

		// Carriage return.
		s.x = 0

		if s.currentLine() == nil {
			// This should _never_ happen, but never say never.
			s.currentLineForWriting()
		}

		// This, and the final line, are the only instances in which newline should
		// be false.
		s.currentLine().newline = false
		s.y++
	}
	// Ensure there are enough lines on screen to start writing here.
	for s.currentLine() == nil {
		// If maxLines is not in use, or adding a new line would not make it
		// larger than maxLines, then just allocate a new line.
		if s.maxLines <= 0 || len(s.screen)+1 <= s.maxLines {
			var nodes []node
			if len(s.nodeRecycling) > 0 {
				// Pop one off the end of nodeRecycling
				r1 := len(s.nodeRecycling) - 1
				nodes = s.nodeRecycling[r1]
				s.nodeRecycling = s.nodeRecycling[:r1]
			}
			if nodes == nil {
				// No slices available for recycling, make a new one.
				nodes = make([]node, 0, s.cols)
			}
			newLine := screenLine{
				nodes:   nodes,
				newline: true,
			}
			s.screen = append(s.screen, newLine)
			if s.y >= s.lines {
				// Because the "window" is always the last s.lines of s.screen
				// (or all of them, if there are fewer lines than s.lines)
				// appending a new line shifts the window down. In that case,
				// compensate by shifting s.y up (eventually to within bounds).
				s.y--
			}
			continue
		}

		// maxLines is in effect, and adding a new line would make the screen
		// larger than maxLines.
		// Pass the whole line being scrolled out to ScrollOutFunc if available,
		// otherwise just scroll out 1 line to nowhere.
		scrollOutTo := 1
		if s.ScrollOutFunc != nil {
			// Whole lines need to be passed to the callback. Find the end of
			// the line (the screen line with newline = true).
			// The majority of the time this will just be the first screen line.
			// If it's all one enormous line, stop at the top of the screen.
			// (so, allow scrollout to eat all of the "scrollback" but none of
			// the "visible screen". We're talking a line that's 160*200
			// chars long for the top of the screen to be reached that way.)
			scrollOutTo = s.top()
			if s.top() == 0 {
				// We still need to scroll out a line, even if there are no lines above
				// the top of the window. Get the next line.
				scrollOutTo = len(s.screen)
			}
			for i, l := range s.screen[:scrollOutTo] {
				if l.newline {
					scrollOutTo = i + 1
					break
				}
			}
			s.ScrollOutFunc(lineToHTML(s.screen[:scrollOutTo]))
		}
		for i := range scrollOutTo {
			s.nodeRecycling = append(s.nodeRecycling, s.screen[i].nodes[:0])
		}
		s.LinesScrolledOut += scrollOutTo

		var nodes []node
		if r1 := len(s.nodeRecycling) - 1; r1 >= 0 {
			// Make a new line on the bottom using a recycled node slice. There's
			// usually at least one we just added.
			nodes = s.nodeRecycling[r1]
			s.nodeRecycling = s.nodeRecycling[:r1]
		} else {
			// No nodes to recycle, make a new node slice. This happens when we scroll
			// out a line that consisted of no screenlines.
			nodes = make([]node, 0, s.cols)
		}
		newLine := screenLine{
			nodes:   nodes,
			newline: true,
		}
		s.screen = append(s.screen[scrollOutTo:], newLine)

		// Since the buffer added 1 line, s.y moves upwards.
		s.y--
	}

	return s.currentLine()
}

// Write a character to the screen's current X&Y, along with the current screen style
func (s *Screen) write(data rune) {
	line := s.currentLineForWriting()
	line.writeNode(s.x, node{blob: data, style: s.style})

	// OSC 8 links work like a style.
	if s.style.hyperlink() {
		if line.hyperlinks == nil {
			line.hyperlinks = make(map[int]string)
		}
		line.hyperlinks[s.x] = s.urlBrush
	}

	s.x++
}

// Append a character to the screen
func (s *Screen) append(data rune) {
	s.write(data)
}

// Append multiple characters to the screen
func (s *Screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

func (s *Screen) appendElement(i *element) {
	line := s.currentLineForWriting()
	idx := len(line.elements)
	line.elements = append(line.elements, i)
	ns := s.style
	ns.setElement(true)

	line.writeNode(s.x, node{blob: rune(idx), style: ns})
	s.x++
}

// Set line metadata. Merges the provided data into any existing
// metadata for the current line, overwriting data when keys collide.
func (s *Screen) setLineMetadata(namespace string, data map[string]string) {
	line := s.currentLineForWriting()
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
	// Wrap slice accesses in a bounds check. Instructions not supplied default
	// to the empty string.
	inst := func(i int) string {
		if i < 0 || i >= len(instructions) {
			return ""
		}
		return instructions[i]
	}

	if strings.HasPrefix(inst(0), "?") {
		// These are typically "private" control sequences, e.g.
		// - show/hide cursor (not relevant)
		// - enable/disable focus reporting (not relevant)
		// - alternate screen buffer (not implemented)
		// - bracketed paste mode (not relevant)
		// Particularly, "show cursor" is CSI ?25h, which would be picked up
		// below if we didn't handle it.
		return
	}

	switch code {
	case 'A': // Cursor Up: go up n
		s.up(inst(0))

	case 'B': // Cursor Down: go down n
		s.down(inst(0))

	case 'C': // Cursor Forward: go right n
		s.forward(inst(0))

	case 'D': // Cursor Back: go left n
		s.backward(inst(0))

	case 'E': // Cursor Next Line: Go to beginning of line n down
		s.x = 0
		s.down(inst(0))

	case 'F': // Cursor Previous Line: Go to beginning of line n up
		s.x = 0
		s.up(inst(0))

	case 'G': // Cursor Horizontal Absolute: Go to column n (default 1)
		s.x = ansiInt(inst(0)) - 1
		s.x = max(s.x, 0)
		s.x = min(s.x, s.cols-1)

	case 'H': // Cursor Position Absolute: Go to row n and column m (default 1;1).
		//
		// There are a variety of agent versions still in use, which have
		// different PTY window settings. Although we emulate a window size
		// here, we can't know for sure which line CSI H is referring to until
		// we have a mechanism to report the real window size that was used.
		// If the program output CSI 1H, we don't know if that's the top of a
		// 80x25 window or a 160x100 window, which could be either 24 lines or
		// 99 lines above the current position.
		//
		// (Relative vertical movement isn't a problem, since the window size
		// only bounds movement; if the program is on an older agent with a
		// smaller window, the relative movement will fit within the larger
		// window emulated here.)
		//
		// Absolute horizontal positioning is much easier. Most programs use
		// CSI G or CSI H to move back to the first column.
		//
		// For now we can pretend that this code is equivalent to "\n" + CSI G
		// (move to an absolute position on the next line). This should be
		// slightly better than the pre-v3.16 behaviour, which completely
		// ignored CSI H, by aligning new content as expected but preserving
		// previous content.
		//
		// Because this "newline" is inserted by us, not the agent, the new line
		// might not have BK metadata (timestamp), so copy it. Also, since the
		// "newline" is only needed to preserve intermediate output, we only
		// need to insert one - multiple CSI H codes without content in between
		// only need one "newline".
		var metadata map[string]string
		if line := s.currentLine(); line != nil && len(line.nodes) > 0 {
			// clone required since setLineMetadata assumes it can own the map
			metadata = maps.Clone(line.metadata[bkNamespace])
			s.y++
		}
		s.x = ansiInt(inst(1)) - 1
		s.x = max(s.x, 0)
		s.x = min(s.x, s.cols-1)
		if metadata != nil {
			s.setLineMetadata(bkNamespace, metadata)
		}

	case 'J': // Erase in Display: Clears part of the screen.
		switch inst(0) {
		case "0", "": // "erase from current position to end (inclusive)"
			s.currentLine().clear(s.x, screenEndOfLine) // same as ESC [0K

			// Rather than truncate s.screen, clear each following line.
			// There's a good chance those lines will be used later, and it
			// avoids having to do maths to fix the cursor position.
			start := s.top() + s.y + 1
			for i := start; i < len(s.screen); i++ {
				s.screen[i].clearAll()
			}

		case "1": // "erase from beginning to current position (inclusive)"
			s.currentLine().clear(screenStartOfLine, s.x) // same as ESC [1K

			// real terms erase part of the window, but the cursor stays still.
			// The intervening lines simply become blank.
			top := s.top()
			end := min(top+s.y, len(s.screen))
			for i := top; i < end; i++ {
				s.screen[i].clearAll()
			}

		case "2":
			// 2: "erase entire display"
			// Previous implementations performed this the same as ESC [3J,
			// which also removes all "scroll-back".
			for i := s.top(); i < len(s.screen); i++ {
				s.screen[i].clearAll()
			}

		case "3":
			// 3: "erase whole display including scroll-back buffer"
			for i := range s.screen {
				s.screen[i].clearAll()
			}
		}

	case 'K': // Erase in Line: erases part of the line.
		switch inst(0) {
		case "0", "":
			s.currentLine().clear(s.x, screenEndOfLine)

		case "1":
			s.currentLine().clear(screenStartOfLine, s.x)

		case "2":
			s.currentLine().clearAll()
		}

	case 'M':
		s.color(instructions)
	}
}

// Write writes ANSI text to the screen.
func (s *Screen) Write(input []byte) (int, error) {
	s.parser.parseToScreen(input)
	return len(input), nil
}

// AsHTML returns the contents of the current screen buffer as HTML.
func (s *Screen) AsHTML() string {
	var sb strings.Builder

	screen := s.screen
	for len(screen) > 0 {
		// Find lineEnd of a line, or failing that, go to the end of the screen.
		lineEnd := len(screen)
		for i, l := range screen {
			if l.newline {
				lineEnd = i + 1
				break
			}
		}
		sb.WriteString(lineToHTML(screen[:lineEnd]))
		screen = screen[lineEnd:]
	}

	// For backwards compatibility the final newline is trimmed.
	return strings.TrimSuffix(sb.String(), "\n")
}

// AsPlainText renders the screen without any ANSI style etc.
func (s *Screen) AsPlainText() string {
	var sb strings.Builder
	for _, line := range s.screen {
		sb.WriteString(line.asPlain())
	}

	// For backwards compatibility the final newline is trimmed.
	return strings.TrimSuffix(sb.String(), "\n")
}

// AsPlainTextWithTimestamps renders the screen as plain text, optionally
// with UTC timestamp prefixes.
func (s *Screen) AsPlainTextWithTimestamps(timestamps bool) string {
	if !timestamps {
		return s.AsPlainText()
	}

	var sb strings.Builder
	for i := 0; i < len(s.screen); {
		// Find the end of this logical line.
		lineEnd := i + 1
		for lineEnd < len(s.screen) && !s.screen[lineEnd-1].newline {
			lineEnd++
		}
		sb.WriteString(lineToPlain(s.screen[i:lineEnd], true))
		i = lineEnd
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

func (s *Screen) newLine() {
	// Do the carriage return first to ensure that currentLineForWriting can't
	// give us the next line if the cursor was placed past the end of the line.
	s.x = 0

	// Ensure the previous line, if it already exists, gets a \n in the render.
	// This could happen if we got CSI A (cursor up), and then \n onto a line
	// that had previously been wrapped from the previous line.
	if line := s.currentLine(); line != nil {
		line.newline = true
	}
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

type screenLine struct {
	nodes []node

	// newline is true for most lines, and means this line ends with \n.
	// If newline is false, this line continues onto the next line.
	newline bool

	// metadata is { namespace => { key => value, ... }, ... }
	// e.g. { "bk" => { "t" => "1234" } }
	metadata map[string]map[string]string

	// element nodes refer to elements in this slice by index
	// (if node.style.element(), then elements[node.blob] is the element)
	elements []*element

	// hyperlinks stores the URL targets for OSC 8 (iTerm-style) links
	// by X position. URLs are too big to fit in every node, most lines won't
	// have links and most nodes in a line won't be linked.
	// So a map is used for sparse storage, only lazily created when text with
	// a link style is written.
	hyperlinks map[int]string
}

func (l *screenLine) clearAll() {
	if l == nil {
		return
	}
	l.nodes = l.nodes[:0]
	l.newline = true
}

// clear clears part (or all) of a line. The range to clear is inclusive
// of xStart and xEnd.
func (l *screenLine) clear(xStart, xEnd int) {
	if l == nil {
		return
	}

	if xStart < 0 {
		xStart = 0
	}
	if xEnd < xStart {
		// Not a valid range.
		return
	}

	if xStart >= len(l.nodes) {
		// Clearing part of a line starting after the end of the current line...
		return
	}

	if xEnd >= len(l.nodes)-1 {
		// Clear from start to end of the line
		l.nodes = l.nodes[:xStart]
		return
	}

	for i := xStart; i <= xEnd; i++ {
		l.nodes[i] = emptyNode
	}
}

func (l *screenLine) writeNode(x int, n node) {
	// Add columns if currently shorter than the cursor's x position
	for i := len(l.nodes); i <= x; i++ {
		l.nodes = append(l.nodes, emptyNode)
	}
	l.nodes[x] = n
}
