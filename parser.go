package terminal

import (
	"unicode"
	"unicode/utf8"
)

const (
	parserModeNormal = iota
	parserModeEscape
	parserModeControl
	parserModeOSC
	parserModeOSCEsc // within OSC and just read an escape
	parserModeCharset
	parserModeAPC
	parserModeAPCEsc // within APC and just read an escape
)

type position struct {
	x, y int
}

// Stateful ANSI parser
type parser struct {
	screen               *Screen
	buffer               join
	remainder            []byte
	mode                 int
	cursor               int
	escapeStartedAt      int
	instructions         []string
	instructionStartedAt int
	savePosition         position

	// Buildkite-specific state
	lastTimestamp int64
}

/*
 * How this state machine works:
 *
 * We start in parserModeNormal. We're not inside an escape sequence. In this mode
 * most input is written directly to the screen. If we receive a newline,
 * backspace or other cursor-moving signal, we let the screen know so that it
 * can change the location of its cursor accordingly.
 *
 * If we're in parserModeNormal and we receive an escape character (\x1b) we enter
 * parserModeEscape. The following character could start an escape sequence, a
 * control sequence, an operating system command, or be invalid or not understood.
 *
 * If we're in parserModeEscape we look for ~~three~~ eight possible characters:
 *
 * 1. For `[` we enter parserModeControl and start looking for a control sequence.
 * 2. For `]` we enter parserModeOSC and look for an operating system command.
 * 3. For `(` or ')' we enter parserModeCharset and look for a character set name.
 * 4. For `_` we enter parserModeAPC and parse the rest of the custom control sequence
 * 5. For `M`, `7`, or `8`, we run an instruction directly (reverse newline,
 *    or save/restore cursor).
 *
 * In all cases we start our instruction buffer. The instruction buffer is used
 * to store the individual characters that make up ANSI instructions before
 * sending them to the screen. If we receive neither of these characters, we
 * treat this as an invalid or unknown escape and return to parserModeNormal.
 *
 * If we're in parserModeControl, we expect to receive a sequence of parameters and
 * then a terminal alphabetic character looking like 1;30;42m. That's an
 * instruction to turn on bold, set the foreground colour to black and the
 * background colour to green. We receive these characters one by one turning
 * the parameters into instruction parts (1, 30, 42) followed by an instruction
 * type (m). Once the instruction type is received we send it and its parts to
 * the screen and return to parserModeNormal.
 *
 * If we're in parserModeOSC, we expect to receive a sequence of characters up to
 * and including a bell (\a) or ESC-\ string terminator. We skip forward until
 * the terminator is reached, then send everything from when we entered parserModeOSC
 * up to the terminator to parseElementSequence and return to parserModeNormal.
 *
 * parserModeAPC is just like parserModeOSC, except the contents should be processed
 * differently.
 *
 * If we're in parserModeCharset we simply discard the next character which would
 * normally designate the character set.
 */

func (p *parser) parseToScreen(input []byte) {

	// This is like append(p.remainder, input), but without copying.
	p.buffer = join{p.remainder, input}

	for p.cursor < p.buffer.len() {
		// UTF-8 runes are 1-4 bytes, so slice ahead +4.
		charBytes := p.buffer.slice(p.cursor, min(p.cursor+4, p.buffer.len()))
		char, charLen := utf8.DecodeRune(charBytes)

		switch p.mode {
		case parserModeEscape:
			// We've received an escape character but aren't inside an escape sequence yet
			p.handleEscape(char)

		case parserModeControl:
			// We're inside a control sequence - figure out its code and its instructions.
			p.handleControlSequence(char)

		case parserModeOSC:
			// We're inside an operating system command, capture until we hit BEL or ESC \ (ST)
			p.handleOperatingSystemCommand(char)

		case parserModeOSCEsc:
			// We're inside an operating system command, and just hit an ESC (might be ST)
			p.handleOSCEscape(char)

		case parserModeCharset:
			// We're inside a charset sequence, capture the next character.
			p.handleCharset(char)

		case parserModeAPC:
			// We're inside a custom escape sequence, capture until we hit BEL or ESC \ (ST)
			p.handleApplicationProgramCommand(char)

		case parserModeAPCEsc:
			// We're inside an APC, and just hit an ESC (which might be ST)
			p.handleAPCEscape(char)

		case parserModeNormal:
			// Outside of an escape sequence entirely, normal input
			p.handleNormal(char)
		}

		p.cursor += charLen
	}

	// If we're in normal mode, everything up to the cursor has been procesed.
	if p.mode == parserModeNormal {
		p.cursor = 0
		p.remainder = p.remainder[:0]
		return
	}

	// We're in the middle of an escape, only everything up to p.escapeStartedAt
	// has been processed. The remainder sits at the end of input, which we
	// don't want to retain (see io.Writer docs), so copy it using append.
	done := p.escapeStartedAt
	p.remainder = append(p.remainder[:0], p.buffer.slice(done, p.buffer.len())...)

	// Adjust the buffer indices accordingly.
	p.cursor -= done
	p.instructionStartedAt -= done
	p.escapeStartedAt -= done
}

// handleCharset is called for each character consumed while in parserModeCharset.
// It ignores the character and transitions back to parserModeNormal.
func (p *parser) handleCharset(rune) {
	p.mode = parserModeNormal
}

// handleOSCEscape is called for the character after an ESC when reading an OSC.
// It either returns to OSC mode, or terminates the OSC and processes it.
func (p *parser) handleOSCEscape(char rune) {
	switch char {
	case '\\': // ESC + \ = string terminator
		// Don't include the ESC in the OSC contents.
		p.processOperatingSystemCommand(p.cursor - 1)

	default:
		// ESC + anything else = not a string terminator.
		// OSC continues...
		p.mode = parserModeOSC
	}
}

// handleOperatingSystemCommand is called for each character consumed while in
// parserModeOSC. It does nothing until the OSC is terminated with either BEL or
// ESC \ (ST).
func (p *parser) handleOperatingSystemCommand(char rune) {
	switch char {
	case '\x07': // BEL terminates the APC
		p.processOperatingSystemCommand(p.cursor)

	case '\x1b': // ESC
		// Next char _could_ be \ which makes the combination a string terminator
		p.mode = parserModeOSCEsc

	default:
		// OSC continues...
	}
}

// processOperatingSystemCommand processes the contents of the OSC that was just read.
func (p *parser) processOperatingSystemCommand(end int) {
	p.mode = parserModeNormal
	element, err := parseElementSequence(string(p.buffer.slice(p.instructionStartedAt, end)))
	// Errors are rendered into the screen (see below).

	if element == nil && err == nil {
		// No element & no error, nothing to render
		return
	}

	ownLine := element == nil || element.elementType == elementImage || element.elementType == elementITermImage

	if ownLine {
		// Images (or the error encountered) should appear on their own line
		if p.screen.x != 0 {
			p.screen.newLine()
		}
		p.screen.currentLine().clear(screenStartOfLine, screenEndOfLine)
	}

	if err != nil {
		p.screen.appendMany([]rune("*** Error parsing custom element escape sequence: "))
		p.screen.appendMany([]rune(err.Error()))
		p.screen.newLine()
		return
	}

	if element != nil && element.elementType == elementITermLink {
		// OSC 8 (iTerm-style) links work like a style. iTerm2 behaves this way.
		// Instead of appending an "element" node, store the URL to apply like a
		// colour. If the URL is empty, the text is no longer linked.
		p.screen.urlBrush = element.url
		p.screen.style.setHyperlink(element.url != "")
		return
	}

	p.screen.appendElement(element)

	if ownLine {
		p.screen.newLine()
	}
}

// handleAPCEscape is called for the character after an ESC when reading an APC.
// It either returns to APC mode, or terminates the APC and processes it.
func (p *parser) handleAPCEscape(char rune) {
	switch char {
	case '\\': // ESC + \ = string terminator
		// Don't include the ESC in the APC contents.
		p.processApplicationProgramCommand(p.cursor - 1)

	default:
		// ESC + anything else = not a string terminator.
		// APC continues...
		p.mode = parserModeAPC
	}
}

// handleApplicationProgramCommand is called for each character consumed while
// in parserModeAPC, but does nothing until the APC is terminated with BEL (0x07)
// or the two-byte form of ST (ESC \).
//
// Technically an APC sequence is terminated by String Terminator (ST; 0x9C or ESC \):
// https://en.wikipedia.org/wiki/C0_and_C1_control_codes#C1_controls
//
// But:
// > For historical reasons, Xterm can end the command with BEL as well as the standard ST
// https://en.wikipedia.org/wiki/ANSI_escape_code#OSC_(Operating_System_Command)_sequences
//
// .. and this is how iTerm2 implements inline images:
// > ESC ] 1337 ; key = value ^G
// https://iterm2.com/documentation-images.html
//
// Buildkite's ansi timestamper does the same, and we don't _expect_ to be
// seeing any other APCs that could be ST-terminated. But we've seen ESC \
// in some bug reports.
func (p *parser) handleApplicationProgramCommand(char rune) {
	switch char {
	case '\x07': // BEL terminates the APC
		p.processApplicationProgramCommand(p.cursor)

	case '\x1b': // ESC
		// Next char _could_ be \ which makes the combination ST
		p.mode = parserModeAPCEsc

	default:
		// APC continues...
	}
}

// processApplicationProgramCommand process the contents of the APC that was just read.
func (p *parser) processApplicationProgramCommand(end int) {
	p.mode = parserModeNormal
	sequence := string(p.buffer.slice(p.instructionStartedAt, end))

	// this might be a Buildkite Application Program Command sequence...
	data, err := p.parseBuildkiteAPC(sequence)
	if err != nil {
		p.screen.appendMany([]rune("*** Error parsing Buildkite APC ANSI escape sequence: "))
		p.screen.appendMany([]rune(err.Error()))
		return
	}

	if data == nil {
		return
	}
	p.screen.setLineMetadata(bkNamespace, data)
}

// handleControlSequence is called for each character consumed while in
// parserModeControl.
func (p *parser) handleControlSequence(char rune) {
	char = unicode.ToUpper(char)
	switch char {
	case '?', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// Part of an instruction

	case ';':
		p.addInstruction()
		p.instructionStartedAt = p.cursor + utf8.RuneLen(';')

	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'Q':
		p.addInstruction()
		p.screen.applyEscape(char, p.instructions)
		p.mode = parserModeNormal

	case 'I', 'L', 'N':
		// CSI i: Enable/disable AUX port
		// CSI L: Set/reset mode (SM/RM)
		// CSI n: Report cursor position
		// All not relevant to us. Swallow the code and continue
		p.mode = parserModeNormal

	default:
		// unrecognized character, abort the escapeCode
		p.cursor = p.escapeStartedAt
		p.mode = parserModeNormal
	}
}

// handleNormal is called for each character consumed while in parserModeNormal.
func (p *parser) handleNormal(char rune) {
	switch char {
	case '\n':
		p.screen.newLine()
	case '\r':
		p.screen.carriageReturn()
	case '\b':
		p.screen.backspace()
	case '\x1b':
		p.escapeStartedAt = p.cursor
		p.mode = parserModeEscape
	default:
		p.screen.append(char)
	}
}

// handleEscape is called for each character consumed while in parserModeEscape.
func (p *parser) handleEscape(char rune) {
	switch char {
	case '[':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.instructions = make([]string, 0, 1)
		p.mode = parserModeControl

	case ']':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.mode = parserModeOSC

	case ')', '(':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('(')
		p.mode = parserModeCharset

	case '_':
		p.instructionStartedAt = p.cursor + utf8.RuneLen('[')
		p.mode = parserModeAPC

	case 'M':
		p.screen.revNewLine()
		p.mode = parserModeNormal

	case '7':
		p.savePosition = position{x: p.screen.x, y: p.screen.y}
		p.mode = parserModeNormal

	case '8':
		p.screen.x = p.savePosition.x
		p.screen.y = p.savePosition.y
		p.mode = parserModeNormal

	case '=', '>': // DECKPAM, DECKPNM
		// These change the keyboard numpad mode between cursor movement
		// and plain digits.
		// For some reason Powershell outputs ESC [?1h ESC =.
		// Swallow and ignore these.
		p.mode = parserModeNormal

	default:
		// Not an escape code, false alarm
		p.cursor = p.escapeStartedAt
		p.mode = parserModeNormal
	}
}

// addInstruction appends an instruction to p.instructions, if the current
// instruction is nonempty.
func (p *parser) addInstruction() {
	instruction := string(p.buffer.slice(p.instructionStartedAt, p.cursor))
	if instruction != "" {
		p.instructions = append(p.instructions, instruction)
	}
}

// join provides a way to slice across consecutive []bytes. Copying happens at
// slice time, not at construction.
type join struct {
	head, tail []byte
}

func (j join) slice(from, to int) []byte {
	m := len(j.head)
	if to <= m {
		return j.head[from:to]
	}
	if from >= m {
		return j.tail[from-m : to-m]
	}
	return append(j.head[from:], j.tail[:to-m]...)
}

func (j join) len() int { return len(j.head) + len(j.tail) }
