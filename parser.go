package terminal

import "unicode"

// Stateful container object for capturing escape codes
type escapeCode struct {
	nextInstruction []rune
	instructions    []string
	buffer          []rune
}

const (
	MODE_NORMAL     = iota
	MODE_PRE_ESCAPE = iota
	MODE_ESCAPE     = iota
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
			// We're inside an escape code - figure out its code and its instructions.
			p.parseEscape(char)
		case MODE_PRE_ESCAPE:
			// We've received an escape character but aren't inside an escape sequence yet
			p.parsePreEscape(char)
		case MODE_NORMAL:
			// Outside of an escape sequence entirely, normal input
			p.parseNormal(char)
		}
	}
}

func (p *parser) parseEscape(char rune) {
	p.escape.buffer = append(p.escape.buffer, char)

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
		p.screen.appendMany([]rune{'\x1b', '['})
		p.screen.appendMany(p.escape.buffer)
		p.mode = MODE_NORMAL
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
		p.mode = MODE_PRE_ESCAPE
	default:
		p.screen.append(char)
	}
}

func (p *parser) parsePreEscape(char rune) {
	if char == '[' {
		p.escape = escapeCode{}
		p.mode = MODE_ESCAPE
	} else {
		// Not an escape code, false alarm
		p.screen.append('\x1b')
		p.parseNormal(char)
		p.mode = MODE_NORMAL
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
