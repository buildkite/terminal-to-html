package terminal

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

type outputBuffer struct {
	buf strings.Builder
}

func (b *outputBuffer) appendNodeStyle(n node) {
	b.buf.WriteString(`<span class="`)
	for idx, class := range n.style.asClasses() {
		if idx > 0 {
			b.buf.WriteString(" ")
		}
		b.buf.WriteString(class)
	}
	b.buf.WriteString(`">`)
}

func (b *outputBuffer) closeStyle() {
	b.buf.WriteString("</span>")
}

func (b *outputBuffer) appendMeta(namespace string, data map[string]string) {
	// We pre-sort the keys to guarantee alphabetical output,
	// because Golang `map`s have guaranteed disorder
	keys := make([]string, len(data))
	// Make a list of the map's keys
	i := 0
	for key := range data {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	b.buf.WriteString("<?" + namespace)
	for i := range keys {
		key := keys[i]
		fmt.Fprintf(&b.buf, ` %s="%s"`, key, html.EscapeString(data[key]))
	}
	b.buf.WriteString("?>")
}

// Append a character to our outputbuffer, escaping HTML bits as necessary.
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

// asHTML returns the line with HTML formatting.
func (l *screenLine) asHTML() string {
	var spanOpen bool
	var lineBuf outputBuffer

	if data, ok := l.metadata[bkNamespace]; ok {
		lineBuf.appendMeta(bkNamespace, data)
	}

	for idx, node := range l.nodes {
		if idx == 0 {
			if !node.style.isPlain() {
				lineBuf.appendNodeStyle(node)
				spanOpen = true
			}
		} else {
			previous := l.nodes[idx-1]
			if !node.hasSameStyle(previous) {
				if spanOpen {
					lineBuf.closeStyle()
					spanOpen = false
				}
				if !node.style.isPlain() {
					lineBuf.appendNodeStyle(node)
					spanOpen = true
				}
			}
		}

		if node.style.element() {
			lineBuf.buf.WriteString(l.elements[node.blob].asHTML())
		} else {
			lineBuf.appendChar(node.blob)
		}
	}
	if spanOpen {
		lineBuf.closeStyle()
	}
	line := strings.TrimRight(lineBuf.buf.String(), " \t")
	if line == "" {
		return "&nbsp;"
	}
	return line
}

// asPlain returns the line contents without any added HTML.
func (l *screenLine) asPlain() string {
	var buf strings.Builder

	for _, node := range l.nodes {
		if !node.style.element() {
			buf.WriteRune(node.blob)
		}
	}

	return strings.TrimRight(buf.String(), " \t")
}
