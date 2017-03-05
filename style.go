package terminal

import "strings"
import "strconv"

var emptyStyle = style{}

type style struct {
	fgColor   uint8
	bgColor   uint8
	fgColorX  bool
	bgColorX  bool
	bold      bool
	faint     bool
	italic    bool
	underline bool
	strike    bool
}

// True if both styles are equal (or are the same object)
func (s *style) isEqual(o *style) bool {
	return s == o || *s == *o
}

// CSS classes that make up the style
func (s *style) asClasses() string {
	if s.isEmpty() {
		return ""
	}

	var styles []string
	if s.fgColor > 0 && s.fgColor < 38 && !s.fgColorX {
		styles = append(styles, "term-fg"+strconv.Itoa(int(s.fgColor)))
	}
	if s.fgColor > 38 && !s.fgColorX {
		styles = append(styles, "term-fgi"+strconv.Itoa(int(s.fgColor)))

	}
	if s.fgColorX {
		styles = append(styles, "term-fgx"+strconv.Itoa(int(s.fgColor)))

	}

	if s.bgColor > 0 && s.bgColor < 48 && !s.bgColorX {
		styles = append(styles, "term-bg"+strconv.Itoa(int(s.bgColor)))
	}
	if s.bgColor > 48 && !s.bgColorX {
		styles = append(styles, "term-bgi"+strconv.Itoa(int(s.bgColor)))
	}
	if s.bgColorX {
		styles = append(styles, "term-bgx"+strconv.Itoa(int(s.bgColor)))
	}

	if s.bold {
		styles = append(styles, "term-fg1")
	}
	if s.faint {
		styles = append(styles, "term-fg2")
	}
	if s.italic {
		styles = append(styles, "term-fg3")
	}
	if s.underline {
		styles = append(styles, "term-fg4")
	}
	if s.strike {
		styles = append(styles, "term-fg9")
	}

	return strings.Join(styles, " ")
}

// True if style is empty
func (s *style) isEmpty() bool {
	return *s == style{}
}

// Add colours to an existing style, potentially returning
// a new style.
func (s *style) color(colors []string) *style {
	if len(colors) == 1 && (colors[0] == "0" || colors[0] == "") {
		// Shortcut for full style reset
		return &emptyStyle
	}

	newStyle := style(*s)
	s = &newStyle

	if len(colors) > 2 {
		cc, err := strconv.ParseUint(colors[2], 10, 8)
		if err != nil {
			return s
		}
		if colors[0] == "38" && colors[1] == "5" {
			// Extended set foreground x-term color
			s.fgColor = uint8(cc)
			s.fgColorX = true
			return s
		}

		// Extended set background x-term color
		if colors[0] == "48" && colors[1] == "5" {
			s.bgColor = uint8(cc)
			s.bgColorX = true
			return s
		}
	}

	for _, ccs := range colors {
		// If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// one, and assign it to s.fgColor or s.bgColor
		cc, err := strconv.ParseUint(ccs, 10, 8)
		if err != nil {
			continue
		}

		switch cc {
		case 0:
			// Reset all styles - don't use &emptyStyle here as we could end up adding colours
			// in this same action.
			s = &style{}
		case 1:
			s.bold = true
			s.faint = false
		case 2:
			s.faint = true
			s.bold = false
		case 3:
			s.italic = true
		case 4:
			s.underline = true
		case 9:
			s.strike = true
		case 21, 22:
			s.bold = false
			s.faint = false
			// Turn off italic
		case 23:
			s.italic = false
			// Turn off underline
		case 24:
			s.underline = false
			// Turn off crossed-out
		case 29:
			s.strike = false
		case 39:
			s.fgColor = 0
			s.fgColorX = false
		case 49:
			s.bgColor = 0
			s.bgColorX = false
			// 30–37, then it's a foreground color
		case 30, 31, 32, 33, 34, 35, 36, 37:
			s.fgColor = uint8(cc)
			s.fgColorX = false
			// 40–47, then it's a background color.
		case 40, 41, 42, 43, 44, 45, 46, 47:
			s.bgColor = uint8(cc)
			s.bgColorX = false
			// 90-97 is like the regular fg color, but high intensity
		case 90, 91, 92, 93, 94, 95, 96, 97:
			s.fgColor = uint8(cc)
			s.fgColorX = false
			// 100-107 is like the regular bg color, but high intensity
		case 100, 101, 102, 103, 104, 105, 106, 107:
			s.bgColor = uint8(cc)
			s.bgColorX = false
		}
	}
	return s
}
