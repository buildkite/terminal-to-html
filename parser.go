package terminal

import (
	"unicode"
	"unicode/utf8"
)

const (
	MODE_NORMAL     = iota
	MODE_PRE_ESCAPE = iota
	MODE_ESCAPE     = iota
)

// Stateful ANSI parser
type parser struct {
	mode                 int
	screen               *screen
	ansi                 []byte
	cursor               int
	escapeStartedAt      int
	instructions         []string
	instructionStartedAt int
}

func parseANSIToScreen(s *screen, ansi []byte) {
	p := parser{mode: MODE_NORMAL, ansi: ansi, screen: s}
	p.mode = MODE_NORMAL
	length := len(p.ansi)
	for p.cursor = 0; p.cursor < length; {
		char, charLen := utf8.DecodeRune(p.ansi[p.cursor:])

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

		p.cursor += charLen
	}
}

func (p *parser) parseEscape(char rune) {
	char = unicode.ToUpper(char)
	switch char {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// Part of an instruction
	case ';':
		p.endOfInstruction()
		p.instructionStartedAt = p.cursor + utf8.RuneLen(';')
	case 'Q', 'K', 'G', 'A', 'B', 'C', 'D', 'M':
		p.endOfInstruction()
		p.screen.applyEscape(char, p.instructions)
		p.mode = MODE_NORMAL
	default:
		// unrecognized character, abort the escapeCode
		p.cursor = p.escapeStartedAt
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
		p.escapeStartedAt = p.cursor
		p.mode = MODE_PRE_ESCAPE
	default:
		p.screen.append(char)
	}
}

func (p *parser) parsePreEscape(char rune) {
	if char == '[' {
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.instructions = make([]string, 0, 1)
		p.mode = MODE_ESCAPE
	} else {
		// Not an escape code, false alarm
		p.cursor = p.escapeStartedAt
		p.mode = MODE_NORMAL
	}
}

// Reset our instruction buffer & add to our instruction list
func (p *parser) endOfInstruction() {
	instruction := string(p.ansi[p.instructionStartedAt:p.cursor])
	p.instructions = append(p.instructions, instruction)
}
