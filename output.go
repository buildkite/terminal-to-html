package terminal

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"strings"
)

type outputBuffer struct {
	buf bytes.Buffer
}

func (b *outputBuffer) appendNodeStyle(n node) {
	b.buf.Write([]byte(`<span class="`))
	for idx, class := range n.style.asClasses() {
		if idx > 0 {
			b.buf.Write([]byte(" "))
		}
		b.buf.Write([]byte(class))
	}
	b.buf.Write([]byte(`">`))
}

func (b *outputBuffer) closeStyle() {
	b.buf.Write([]byte("</span>"))
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

func outputLineAsHTML(line screenLine) string {
	var spanOpen bool
	var lineBuf outputBuffer

	if data, ok := line.metadata[bkNamespace]; ok {
		lineBuf.appendMeta(bkNamespace, data)
	}

	for idx, node := range line.nodes {
		if idx == 0 && !node.style.isEmpty() {
			lineBuf.appendNodeStyle(node)
			spanOpen = true
		} else if idx > 0 {
			previous := line.nodes[idx-1]
			if !node.hasSameStyle(previous) {
				if spanOpen {
					lineBuf.closeStyle()
					spanOpen = false
				}
				if !node.style.isEmpty() {
					lineBuf.appendNodeStyle(node)
					spanOpen = true
				}
			}
		}

		if elem := node.elem; elem != nil {
			lineBuf.buf.WriteString(elem.asHTML())
		}

		if r, ok := node.getRune(); ok {
			lineBuf.appendChar(r)
		}
	}
	if spanOpen {
		lineBuf.closeStyle()
	}
	return strings.TrimRight(lineBuf.buf.String(), " \t")
}
