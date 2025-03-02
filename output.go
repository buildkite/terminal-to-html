package terminal

import (
	"html"
	"html/template"
	"maps"
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
	strings.Builder
}

func (b *outputBuffer) appendNodeStyle(n node) {
	openSpanTagTmpl.Execute(b, strings.Join(n.style.asClasses(), " "))
}

func (b *outputBuffer) closeStyle() {
	b.WriteString("</span>")
}

func (b *outputBuffer) appendAnchor(url string) {
	b.WriteString(`<a href="`)
	b.WriteString(html.EscapeString(sanitizeURL(url)))
	b.WriteString(`">`)
}

func (b *outputBuffer) closeAnchor() {
	b.WriteString("</a>")
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
	timeTagImpl.Execute(b, datetime)
}

// Append a character to our outputbuffer, escaping HTML bits as necessary.
func (b *outputBuffer) appendChar(char rune) {
	switch char {
	case '&':
		b.WriteString("&amp;")
	case '\'':
		b.WriteString("&#39;")
	case '<':
		b.WriteString("&lt;")
	case '>':
		b.WriteString("&gt;")
	case '"':
		b.WriteString("&quot;")
	case '/':
		b.WriteString("&#47;")
	default:
		b.WriteRune(char)
	}
}

// lineToHTML joins parts of a line together and renders them in HTML. It
// ignores the newline field (i.e. assumes all parts are !newline except the
// last part). The output string will have a terminating \n.
func lineToHTML(parts []screenLine) string {
	var buf outputBuffer

	// Combine metadata - last metadata wins.
	bkmd := make(map[string]string)
	for _, l := range parts {
		maps.Copy(bkmd, l.metadata[bkNamespace])
	}
	if len(bkmd) > 0 {
		buf.appendMeta(bkNamespace, bkmd)
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
				buf.closeAnchor()
			case tagSpan:
				buf.closeStyle()
			}
		}
		tagStack = tagStack[:idx]
	}

	// The zero value for node has a plain style and no hyperlink.
	var previous node

	for _, l := range parts {
		for x, current := range l.nodes {
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
				buf.appendAnchor(l.hyperlinks[x])
				tagStack = append(tagStack, tagAnchor)
			}
			// Open a new span tag, if one is not already open and this node has
			// style.
			if !slices.Contains(tagStack, tagSpan) && !current.style.isPlain() {
				buf.appendNodeStyle(current)
				tagStack = append(tagStack, tagSpan)
			}

			// Write a standalone element or a rune.
			if current.style.element() {
				buf.WriteString(l.elements[current.blob].asHTML())
			} else {
				buf.appendChar(current.blob)
			}

			previous = current
		}
	}

	// Close any that are open, in reverse order that they were opened.
	closeFrom(0)

	out := strings.TrimRight(buf.String(), " \t")
	if out == "" {
		return "&nbsp;\n"
	}
	out += "\n"
	return out
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
