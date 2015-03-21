package terminal

import "unicode"

// Stateful container object for capturing escape codes
type escapeCode struct {
	nextInstruction []rune
	instructions    []string
	buffer          []rune
}

const (
	MODE_NORMAL = iota
	MODE_ESCAPE = iota
)

// Stateful ANSI parser
type parser struct {
	mode   int
	escape escapeCode
	screen *screen
}

func (p *parser) parse(ansi []byte) {
	p.mode = MODE_NORMAL
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
		// Expect our second character to be [, abort otherwise
		if char != '[' {
			p.abortEscape()
		}
		return
	}

	char = unicode.ToUpper(char)
	switch char {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.escape.nextInstruction = append(p.escape.nextInstruction, char)
	case ';':
		p.escape.endOfInstruction()
	case 'Q', 'K', 'G', 'A', 'B', 'C', 'D', 'M':
		p.escape.endOfInstruction()
		p.screen.applyEscape(char, p.escape.instructions)
		p.mode = MODE_NORMAL
	default:
		// unrecognized character, abort the escapeCode
		p.abortEscape()
	}
}

// Abort an escape code, blat what we have back to the screen
func (p *parser) abortEscape() {
	p.screen.appendMany(p.escape.buffer)
	p.mode = MODE_NORMAL
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
