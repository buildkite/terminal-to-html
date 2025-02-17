package terminal

import (
	"html"
	"html/template"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	timeTagImpl = template.Must(template.New("time").Parse(
		`<time datetime="{{.}}">{{.}}</time>`,
	))

	openSpanTagTmpl = template.Must(template.New("span").Parse(
		`<span class="{{.}}">`,
	))
)

type outputBuffer struct {
	buf strings.Builder
}

func (b *outputBuffer) appendNodeStyle(n node) {
	openSpanTagTmpl.Execute(&b.buf, strings.Join(n.style.asClasses(), " "))
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
	// We only support the bk namespace and a well-formed millisecond epoch.
	if namespace != bkNamespace {
		return
	}
	millis, err := strconv.ParseInt(data["t"], 10, 64)
	if err != nil {
		return
	}
	time := time.Unix(millis/1000, (millis%1000)*1_000_000).UTC()
	// One of the formats accepted by the <time> tag:
	datetime := time.Format("2006-01-02T15:04:05.999Z")
	timeTagImpl.Execute(&b.buf, datetime)
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
func (l *screenLine) asHTML(allowMetadata bool) string {
	var lineBuf outputBuffer

	if data, ok := l.metadata[bkNamespace]; ok && allowMetadata {
		lineBuf.appendMeta(bkNamespace, data)
	}

	// tagStack is used as a stack of open tags, so they can be closed in the
	// right order. We only have two kinds of tag, so the stack should be tiny,
	// but the algorithm can be extended later if needed.
	tagStack := make([]int, 0, 2)
	const (
		tagAnchor = iota
		tagSpan
	)

	// Close tags in the stack, starting at idx. They're closed in the reverse
	// order they were opened.
	closeFrom := func(idx int) {
		for i := len(tagStack) - 1; i >= idx; i-- {
			switch tagStack[i] {
			case tagAnchor:
				lineBuf.closeAnchor()
			case tagSpan:
				lineBuf.closeStyle()
			}
		}
		tagStack = tagStack[:idx]
	}

	for x, current := range l.nodes {
		// The zero value for node has a plain style and no hyperlink.
		var previous node
		// If we're past the first node in the line, there is a previous node
		// to the left.
		if x > 0 {
			previous = l.nodes[x-1]
		}

		// A set of flags for which tags need changing.
		tagChanged := []bool{
			// The anchor tag needs changing if the link "style" has changed,
			// or if they are both links the link URLs are different.
			// (Note that the x-1 index into the hyperlinks map returns "".)
			tagAnchor: current.style.hyperlink() != previous.style.hyperlink() ||
				(current.style.hyperlink() && l.hyperlinks[x-1] != l.hyperlinks[x]),

			// The span tag needs changing if the style has changed.
			tagSpan: !current.hasSameStyle(previous),
		}

		// Go forward through the stack of open tags, looking for the first
		// tag we need to close (because it changed).
		// If none are found, closeFromIdx will be past the end of the stack.
		closeFromIdx := len(tagStack)
		for i, ot := range tagStack {
			if tagChanged[ot] {
				closeFromIdx = i
				break
			}
		}

		// Close everything from that stack index onwards.
		closeFrom(closeFromIdx)

		// Now open new tags as needed.
		// Open a new anchor tag, if one is not already open and this node is
		// hyperlinked.
		if !slices.Contains(tagStack, tagAnchor) && current.style.hyperlink() {
			lineBuf.appendAnchor(l.hyperlinks[x])
			tagStack = append(tagStack, tagAnchor)
		}
		// Open a new span tag, if one is not already open and this node has
		// style.
		if !slices.Contains(tagStack, tagSpan) && !current.style.isPlain() {
			lineBuf.appendNodeStyle(current)
			tagStack = append(tagStack, tagSpan)
		}

		// Write a standalone element or a rune.
		if current.style.element() {
			lineBuf.buf.WriteString(l.elements[current.blob].asHTML())
		} else {
			lineBuf.appendChar(current.blob)
		}
	}

	// Close any that are open, in reverse order that they were opened.
	closeFrom(0)

	line := lineBuf.buf.String()
	if l.newline {
		line = strings.TrimRight(line, " \t")
	}
	if line == "" {
		line = "&nbsp;"
	}
	if l.newline {
		line += "\n"
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

	line := buf.String()
	if l.newline {
		line = strings.TrimRight(line, " \t") + "\n"
	}
	return line
}
