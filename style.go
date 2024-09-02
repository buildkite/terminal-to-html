package terminal

import "strconv"

type style uint32

// style encoding:
// 0...  ...7  8... ...15  16...23     24     25   26....31
// [fg color]  [bg color]  [flags]  element  link  [unused]
// flags = bold, faint, etc

const (
	sbFGColorX = 1 << (16 + iota)
	sbBGColorX
	sbBold
	sbFaint
	sbItalic
	sbUnderline
	sbStrike
	sbBlink
	sbElement   // meaning: this node is actually an element
	sbHyperlink // this node is styled with an OSC 8 (iTerm-style) link
)

// Used for comparing styles - ignores the element bit, link bit, and unused bits.
const styleComparisonMask = 0x00ffffff

// isPlain reports if there is no style information. elements (that have no
// other style set) are also considered plain.
func (s style) isPlain() bool { return s&styleComparisonMask == 0 }

func (s style) fgColor() uint8  { return uint8(s & 0xff) }
func (s style) bgColor() uint8  { return uint8((s & 0xff_00) >> 8) }
func (s style) fgColorX() bool  { return s&sbFGColorX != 0 }
func (s style) bgColorX() bool  { return s&sbBGColorX != 0 }
func (s style) bold() bool      { return s&sbBold != 0 }
func (s style) faint() bool     { return s&sbFaint != 0 }
func (s style) italic() bool    { return s&sbItalic != 0 }
func (s style) underline() bool { return s&sbUnderline != 0 }
func (s style) strike() bool    { return s&sbStrike != 0 }
func (s style) blink() bool     { return s&sbBlink != 0 }
func (s style) element() bool   { return s&sbElement != 0 }
func (s style) hyperlink() bool { return s&sbHyperlink != 0 }

func (s *style) setFGColor(v uint8)  { *s = (*s &^ 0xff) | style(v) }
func (s *style) setBGColor(v uint8)  { *s = (*s &^ 0xff_00) | (style(v) << 8) }
func (s *style) setFGColorX(v bool)  { *s = (*s &^ sbFGColorX) | booln(v, sbFGColorX) }
func (s *style) setBGColorX(v bool)  { *s = (*s &^ sbBGColorX) | booln(v, sbBGColorX) }
func (s *style) setBold(v bool)      { *s = (*s &^ sbBold) | booln(v, sbBold) }
func (s *style) setFaint(v bool)     { *s = (*s &^ sbFaint) | booln(v, sbFaint) }
func (s *style) setItalic(v bool)    { *s = (*s &^ sbItalic) | booln(v, sbItalic) }
func (s *style) setUnderline(v bool) { *s = (*s &^ sbUnderline) | booln(v, sbUnderline) }
func (s *style) setStrike(v bool)    { *s = (*s &^ sbStrike) | booln(v, sbStrike) }
func (s *style) setBlink(v bool)     { *s = (*s &^ sbBlink) | booln(v, sbBlink) }
func (s *style) setElement(v bool)   { *s = (*s &^ sbElement) | booln(v, sbElement) }
func (s *style) setHyperlink(v bool) { *s = (*s &^ sbHyperlink) | booln(v, sbHyperlink) }

const (
	COLOR_NORMAL        = iota
	COLOR_GOT_38_NEED_5 = iota
	COLOR_GOT_48_NEED_5 = iota
	COLOR_GOT_38        = iota
	COLOR_GOT_48        = iota
)

// CSS classes that make up the style
func (s style) asClasses() []string {
	var styles []string

	if s.fgColor() > 0 && s.fgColor() < 38 && !s.fgColorX() {
		styles = append(styles, "term-fg"+strconv.Itoa(int(s.fgColor())))
	}
	if s.fgColor() > 38 && !s.fgColorX() {
		styles = append(styles, "term-fgi"+strconv.Itoa(int(s.fgColor())))

	}
	if s.fgColorX() {
		styles = append(styles, "term-fgx"+strconv.Itoa(int(s.fgColor())))

	}

	if s.bgColor() > 0 && s.bgColor() < 48 && !s.bgColorX() {
		styles = append(styles, "term-bg"+strconv.Itoa(int(s.bgColor())))
	}
	if s.bgColor() > 48 && !s.bgColorX() {
		styles = append(styles, "term-bgi"+strconv.Itoa(int(s.bgColor())))
	}
	if s.bgColorX() {
		styles = append(styles, "term-bgx"+strconv.Itoa(int(s.bgColor())))
	}

	if s.bold() {
		styles = append(styles, "term-fg1")
	}
	if s.faint() {
		styles = append(styles, "term-fg2")
	}
	if s.italic() {
		styles = append(styles, "term-fg3")
	}
	if s.underline() {
		styles = append(styles, "term-fg4")
	}
	if s.blink() {
		styles = append(styles, "term-fg5")
	}
	if s.strike() {
		styles = append(styles, "term-fg9")
	}

	return styles
}

// Add colours to an existing style, returning a new style.
func (s style) color(colors []string) style {
	if len(colors) == 0 || (len(colors) == 1 && (colors[0] == "0" || colors[0] == "")) {
		// s with all normal styles masked out
		return s &^ styleComparisonMask
	}

	colorMode := COLOR_NORMAL

	for _, ccs := range colors {
		// If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// one, and assign it to s.fgColor or s.bgColor
		cc, err := strconv.ParseUint(ccs, 10, 8)
		if err != nil {
			continue
		}

		// State machine for XTerm colors, eg 38;5;150
		switch colorMode {
		case COLOR_GOT_38_NEED_5:
			if cc == 5 {
				colorMode = COLOR_GOT_38
			} else {
				colorMode = COLOR_NORMAL
			}
			continue
		case COLOR_GOT_48_NEED_5:
			if cc == 5 {
				colorMode = COLOR_GOT_48
			} else {
				colorMode = COLOR_NORMAL
			}
			continue
		case COLOR_GOT_38:
			s.setFGColor(uint8(cc))
			s.setFGColorX(true)
			colorMode = COLOR_NORMAL
			continue
		case COLOR_GOT_48:
			s.setBGColor(uint8(cc))
			s.setBGColorX(true)
			colorMode = COLOR_NORMAL
			continue
		}

		switch cc {
		case 0:
			// Reset all styles
			s &^= styleComparisonMask
		case 1:
			s.setBold(true)
			s.setFaint(false)
		case 2:
			s.setFaint(true)
			s.setBold(false)
		case 3:
			s.setItalic(true)
		case 4:
			s.setUnderline(true)
		case 5, 6:
			s.setBlink(true)
		case 9:
			s.setStrike(true)
		case 21, 22:
			s.setBold(false)
			s.setFaint(false)
		case 23:
			s.setItalic(false)
		case 24:
			s.setUnderline(false)
		case 25:
			s.setBlink(false)
		case 29:
			s.setStrike(false)
		case 38:
			colorMode = COLOR_GOT_38_NEED_5
		case 39:
			s.setFGColor(0)
			s.setFGColorX(false)
		case 48:
			colorMode = COLOR_GOT_48_NEED_5
		case 49:
			s.setBGColor(0)
			s.setBGColorX(false)
		case 30, 31, 32, 33, 34, 35, 36, 37, 90, 91, 92, 93, 94, 95, 96, 97:
			s.setFGColor(uint8(cc))
			s.setFGColorX(false)
		case 40, 41, 42, 43, 44, 45, 46, 47, 100, 101, 102, 103, 104, 105, 106, 107:
			s.setBGColor(uint8(cc))
			s.setBGColorX(false)
		}
	}
	return s
}

// false, true => 0, t
func booln(b bool, t style) style {
	if b {
		return t
	}
	return 0
}
