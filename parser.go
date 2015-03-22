package terminal

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MODE_NORMAL       = iota
	MODE_PRE_ESCAPE   = iota
	MODE_ESCAPE       = iota
	MODE_ITERM_ESCAPE = iota
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
		case MODE_ITERM_ESCAPE:
			// We're inside an iTerm escape sequence, capture until we hit a bell character
			p.parseItermEscape(char)
		case MODE_NORMAL:
			// Outside of an escape sequence entirely, normal input
			p.parseNormal(char)
		}

		p.cursor += charLen
	}
}

func (p *parser) parseItermEscape(char rune) {
	if char != '\a' {
		return
	}

	// Bell received, stop parsing our potential image
	itermImage, err := parseItermImageSequence(string(p.ansi[p.instructionStartedAt:p.cursor]))

	// Images (or the error encountered) should appear on their own line
	if p.screen.x != 0 {
		p.screen.newLine()
	}
	p.screen.clear(p.screen.y, screenStartOfLine, screenEndOfLine)

	if err != nil {
		p.screen.appendMany([]rune("*** Error parsing iTerm2 image escape sequence: "))
		p.screen.appendMany([]rune(err.Error()))
	} else {
		p.screen.appendImage(itermImage)
	}
	p.screen.newLine()

	p.mode = MODE_NORMAL
}

type itermImage struct {
	alt          string
	content_type string
	content      string
	height       string
	width        string
}

func (i *itermImage) asHTML() string {
	return fmt.Sprintf(`<img alt=%q src="data:%s;base64,%s">`, i.alt, i.content_type, i.content)
}

func parseItermImageSequence(sequence string) (*itermImage, error) {
	// Expect 1337;File=name=1.gif;inline=1:BASE64

	imageInline := false

	prefixLen := len("1337;File=")
	if !strings.HasPrefix(sequence, "1337;File=") {
		if len(sequence) > prefixLen {
			sequence = sequence[:prefixLen] // Don't blow out our error output
		}
		return nil, fmt.Errorf("Expected sequence to start with 1337;File=, got %q instead", sequence)
	}
	sequence = sequence[prefixLen:]

	parts := strings.Split(sequence, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Expected sequence to have one arguments part and one content part, got %d parts", len(parts))
	}
	arguments := parts[0]
	content := parts[1]

	_, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("Expected content part to be valid Base64")
	}

	img := &itermImage{content: content}
	argsSplit := strings.Split(arguments, ";")
	for _, arg := range argsSplit {
		argParts := strings.SplitN(arg, "=", 2)
		if len(argParts) != 2 {
			continue
		}
		key := argParts[0]
		val := argParts[1]
		switch strings.ToLower(key) {
		case "name":
			img.alt = val
			img.content_type = contentTypeForFile(val)
		case "inline":
			imageInline = val == "1"
		}
	}

	if img.alt == "" {
		return nil, fmt.Errorf("name= argument not supplied, required to determine content type")
	}

	if !imageInline {
		// in iTerm2, if you don't specify inline=1, the image is merely downloaded
		// and not displayed.
		img = nil
	}
	return img, nil
}

func contentTypeForFile(filename string) string {
	return "image/gif"
}

func (p *parser) parseEscape(char rune) {
	char = unicode.ToUpper(char)
	switch char {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// Part of an instruction
	case ';':
		p.addInstruction()
		p.instructionStartedAt = p.cursor + utf8.RuneLen(';')
	case 'Q', 'K', 'G', 'A', 'B', 'C', 'D', 'M':
		p.addInstruction()
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
		p.screen.newLine()
	case '\r':
		p.screen.carriageReturn()
	case '\b':
		p.screen.backspace()
	case '\x1b':
		p.escapeStartedAt = p.cursor
		p.mode = MODE_PRE_ESCAPE
	default:
		p.screen.append(char)
	}
}

func (p *parser) parsePreEscape(char rune) {
	switch char {
	case '[':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.instructions = make([]string, 0, 1)
		p.mode = MODE_ESCAPE
	case ']':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.mode = MODE_ITERM_ESCAPE
	default:
		// Not an escape code, false alarm
		p.cursor = p.escapeStartedAt
		p.mode = MODE_NORMAL
	}
}

func (p *parser) addInstruction() {
	instruction := string(p.ansi[p.instructionStartedAt:p.cursor])
	p.instructions = append(p.instructions, instruction)
}
