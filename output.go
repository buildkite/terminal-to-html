package terminal

import (
	"fmt"
	"html"
	"slices"
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

func (b *outputBuffer) appendAnchor(url string) {
	b.buf.WriteString(`<a href="`)
	b.buf.WriteString(html.EscapeString(sanitizeURL(url)))
	b.buf.WriteString(`">`)
}

func (b *outputBuffer) closeAnchor() {
	b.buf.WriteString("</a>")
}

func (b *outputBuffer) appendMeta(namespace string, data map[string]string) {
	// We pre-sort the keys to guarantee alphabetical output,
	// because Go's maps have guaranteed disorder.
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	b.buf.WriteString("<?" + namespace)
	for _, key := range keys {
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
	var spanOpen, anchorOpen bool
	var lineBuf outputBuffer

	if data, ok := l.metadata[bkNamespace]; ok {
		lineBuf.appendMeta(bkNamespace, data)
	}

	for x, node := range l.nodes {
		// First node on line?
		if x == 0 {
			// Open anchors before spans, as needed.
			if node.style.hyperlink() {
				lineBuf.appendAnchor(l.hyperlinks[x])
				anchorOpen = true
			}
			if !node.style.isPlain() {
				lineBuf.appendNodeStyle(node)
				spanOpen = true
			}
		} else { // not the first node
			// Close span tags before closing anchor tags.
			previous := l.nodes[x-1]
			sameStyle := node.hasSameStyle(previous)
			if !sameStyle && spanOpen {
				lineBuf.closeStyle()
				spanOpen = false
			}

			sameLink := node.style.hyperlink() == previous.style.hyperlink()
			if sameLink && node.style.hyperlink() {
				sameLink = l.hyperlinks[x-1] == l.hyperlinks[x]
			}
			if !sameLink {
				// Close the old anchor tag.
				if anchorOpen {
					lineBuf.closeAnchor()
					anchorOpen = false
				}
				// Open a new anchor tag (if this node is hyperlinked).
				if node.style.hyperlink() {
					lineBuf.appendAnchor(l.hyperlinks[x])
					anchorOpen = true
				}
			}
			// Open a new span tag if this node has style.
			if !sameStyle && !node.style.isPlain() {
				lineBuf.appendNodeStyle(node)
				spanOpen = true
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
	if anchorOpen {
		lineBuf.closeAnchor()
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
