package terminal

import (
	"bytes"
	"strings"
)

func (n *node) hasSameStyle(o node) bool {
	return n == &o || n.style.string() == o.style.string()
}

type outputBuffer struct {
	buf bytes.Buffer
}

func (b *outputBuffer) appendNodeStyle(n node) {
	b.buf.Write([]byte(`<span class="`))
	b.buf.Write([]byte(n.style.string()))
	b.buf.Write([]byte(`">`))
}

func (b *outputBuffer) closeStyle() {
	b.buf.Write([]byte("</span>"))
}

func (s *screen) output() []byte {
	var lines []string

	for _, line := range s.screen {
		var openStyles int
		var lineBuf outputBuffer

		for idx, node := range line {
			if idx == 0 && !node.style.empty() {
				lineBuf.appendNodeStyle(node)
				openStyles++
			} else if idx > 0 {
				previous := line[idx-1]
				if !node.hasSameStyle(previous) {
					if node.style.empty() {
						lineBuf.closeStyle()
						openStyles--
					} else {
						lineBuf.appendNodeStyle(node)
						openStyles++
					}
				}
			}
			lineBuf.appendChar(node.blob)
		}
		for i := 0; i < openStyles; i++ {
			lineBuf.closeStyle()
		}
		asString := strings.TrimRight(lineBuf.buf.String(), " \t")

		lines = append(lines, asString)
	}

	return []byte(strings.Join(lines, "\n"))
}

func (b *outputBuffer) appendChar(char rune) {
	switch char {
	case '&':
		b.buf.WriteString("&amp;")
	case '\'':
		b.buf.WriteString("&#39;")
	case '<':
		b.buf.WriteString("&lt;")
	case '>':
		b.buf.WriteString("&gt;")
	case '"':
		b.buf.WriteString("&quot;")
	case '/':
		b.buf.WriteString("&#47;")
	default:
		b.buf.WriteRune(char)
	}
}
