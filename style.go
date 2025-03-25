package terminal

import "strconv"

type style uint64

// style encoding:
// 0... ...23  24... ...47  48...57     58     59   60....63
// [fg color]  [bg color]   [flags]  element  link  [unused]
// flags = bold, faint, etc

const (
	sbFGColorX1 = style(1) << (48 + iota)
	sbFGColorX2
	sbBGColorX1
	sbBGColorX2
	sbBold
	sbFaint
	sbItalic
	sbUnderline
	sbStrike
	sbBlink
	sbElement   // meaning: this node is actually an element
	sbHyperlink // this node is styled with an OSC 8 (iTerm-style) link
)

const (
	sbFGColorX = sbFGColorX1 | sbFGColorX2
	sbBGColorX = sbBGColorX1 | sbBGColorX2
)

const (
	colorNone = uint8(iota)
	colorSGR
	color8Bit
	color24Bit
)


// Used for comparing styles - ignores the element bit, link bit, and unused bits.
const styleComparisonMask = 0x03ff_ffff_ffff_ffff

// isPlain reports if there is no style information. elements (that have no
// other style set) are also considered plain.
func (s style) isPlain() bool { return s&styleComparisonMask == 0 }

func (s style) fgColor() uint32  { return uint32(s & 0x0000_00ff_ffff) }
func (s style) fgColorType() uint8  { return uint8((s&sbFGColorX) >> 48) }
func (s style) bgColor() uint32  { return uint32((s & 0xffff_ff00_0000) >> 24) }
func (s style) bgColorType() uint8  { return uint8((s&sbBGColorX) >> 50) }
func (s style) bold() bool       { return s&sbBold != 0 }
func (s style) faint() bool      { return s&sbFaint != 0 }
func (s style) italic() bool     { return s&sbItalic != 0 }
func (s style) underline() bool  { return s&sbUnderline != 0 }
func (s style) strike() bool     { return s&sbStrike != 0 }
func (s style) blink() bool      { return s&sbBlink != 0 }
func (s style) element() bool    { return s&sbElement != 0 }
func (s style) hyperlink() bool  { return s&sbHyperlink != 0 }

func (s *style) resetFGColor()  { *s = (*s &^ 0x3_0000_00ff_ffff) }
func (s *style) setFGColorSGR(v uint8)  { *s = (*s &^ 0x3_0000_00ff_ffff) | style(v) | (style(colorSGR) << 48) }
func (s *style) setFGColor8Bit(v uint8)  { *s = (*s &^ 0x3_0000_00ff_ffff) | style(v) | (style(color8Bit) << 48) }
func (s *style) setFGColor24Bit(rgb [3]uint8)  { *s = (*s &^ 0x3_0000_00ff_ffff) | (style(rgb[0]) << 16) | (style(rgb[1]) << 8) | style(rgb[2]) | (style(color24Bit) << 48) }

func (s *style) resetBGColor()  { *s = (*s &^ 0xc_ffff_ff00_0000) }
func (s *style) setBGColorSGR(v uint8)  { *s = (*s &^ 0xc_ffff_ff00_0000) | (style(v) << 24) | (style(colorSGR) << 50) }
func (s *style) setBGColor8Bit(v uint8)  { *s = (*s &^ 0xc_ffff_ff00_0000) | (style(v) << 24) | (style(color8Bit) << 50) }
func (s *style) setBGColor24Bit(rgb [3]uint8)  { *s = (*s &^ 0xc_ffff_ff00_0000) | (style(rgb[0]) << 40) | (style(rgb[1]) << 32) | (style(rgb[2]) << 24) | (style(color24Bit) << 50) }

func (s *style) setBold(v bool)      { *s = (*s &^ sbBold) | booln(v, sbBold) }
func (s *style) setFaint(v bool)     { *s = (*s &^ sbFaint) | booln(v, sbFaint) }
func (s *style) setItalic(v bool)    { *s = (*s &^ sbItalic) | booln(v, sbItalic) }
func (s *style) setUnderline(v bool) { *s = (*s &^ sbUnderline) | booln(v, sbUnderline) }
func (s *style) setStrike(v bool)    { *s = (*s &^ sbStrike) | booln(v, sbStrike) }
func (s *style) setBlink(v bool)     { *s = (*s &^ sbBlink) | booln(v, sbBlink) }
func (s *style) setElement(v bool)   { *s = (*s &^ sbElement) | booln(v, sbElement) }
func (s *style) setHyperlink(v bool) { *s = (*s &^ sbHyperlink) | booln(v, sbHyperlink) }

const (
	COLOR_NORMAL   = iota
	COLOR_GOT_38_2 = iota
	COLOR_GOT_38_5 = iota
	COLOR_GOT_48_2 = iota
	COLOR_GOT_48_5 = iota
	COLOR_GOT_38   = iota
	COLOR_GOT_48   = iota
)

// CSS classes that make up the style
func (s style) asClasses() []string {
	var styles []string

	switch s.fgColorType() {
	case colorSGR:
		if s.fgColor() > 29 && s.fgColor() < 38 {
			styles = append(styles, "term-fg"+strconv.Itoa(int(s.fgColor())))
		}
		if s.fgColor() > 89 && s.fgColor() < 98 {
			styles = append(styles, "term-fgi"+strconv.Itoa(int(s.fgColor())))
		}
	case color8Bit:
		styles = append(styles, "term-fgx"+strconv.Itoa(int(s.fgColor())))
	case color24Bit:
		// This should set the fg color to the specified color, but the previous behavior was
		// undefined and implementing that is outside the scope of this PR.
	}

	switch s.bgColorType() {
	case colorSGR:
		if s.bgColor() > 39 && s.bgColor() < 48 {
			styles = append(styles, "term-bg"+strconv.Itoa(int(s.bgColor())))
		}
		if s.bgColor() > 99 && s.bgColor() < 108 {
			styles = append(styles, "term-bgi"+strconv.Itoa(int(s.bgColor())))
		}
	case color8Bit:
		styles = append(styles, "term-bgx"+strconv.Itoa(int(s.bgColor())))
	case color24Bit:
		// This should set the bg color to the specified color, but the previous behavior was
		// undefined and implementing that is outside the scope of this PR.
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
	var rgb [3]uint8
	var rgb_index uint8

	for _, ccs := range colors {
		// If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// one, and assign it to s.fgColor or s.bgColor
		cc, err := strconv.ParseUint(ccs, 10, 8)
		if err != nil {
			continue
		}

		// State machine for XTerm colors, eg 38;5;150
		switch colorMode {
		case COLOR_GOT_38:
			switch cc {
			case 5:
				colorMode = COLOR_GOT_38_5
			case 2:
				colorMode = COLOR_GOT_38_2
				rgb_index = 0
			default:
				colorMode = COLOR_NORMAL
			}
			continue
		case COLOR_GOT_48:
			switch cc {
			case 5:
				colorMode = COLOR_GOT_48_5
			case 2:
				colorMode = COLOR_GOT_48_2
				rgb_index = 0
			default:
				colorMode = COLOR_NORMAL
			}
			continue
		case COLOR_GOT_38_5:
			s.setFGColor8Bit(uint8(cc))
			colorMode = COLOR_NORMAL
			continue
		case COLOR_GOT_48_5:
			s.setBGColor8Bit(uint8(cc))
			colorMode = COLOR_NORMAL
			continue
		case COLOR_GOT_38_2:
			rgb[rgb_index] = uint8(cc)
			if rgb_index == 3 {
				s.setFGColor24Bit(rgb)
				colorMode = COLOR_NORMAL
				continue
			}
			rgb_index++
			continue
		case COLOR_GOT_48_2:
			rgb[rgb_index] = uint8(cc)
			if rgb_index == 3 {
				s.setBGColor24Bit(rgb)
				colorMode = COLOR_NORMAL
				continue
			}
			rgb_index++
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
			colorMode = COLOR_GOT_38
		case 39:
			s.resetFGColor()
		case 48:
			colorMode = COLOR_GOT_48
		case 49:
			s.resetBGColor()
		case 30, 31, 32, 33, 34, 35, 36, 37, 90, 91, 92, 93, 94, 95, 96, 97:
			s.setFGColorSGR(uint8(cc))
		case 40, 41, 42, 43, 44, 45, 46, 47, 100, 101, 102, 103, 104, 105, 106, 107:
			s.setBGColorSGR(uint8(cc))
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
