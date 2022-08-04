package terminal

var emptyNode = node{blob: ' ', style: &emptyStyle}

type node struct {
	blob  rune
	style *style
	elem  *element
}

func (n *node) hasSameStyle(o node) bool {
	return n.style.isEqual(o.style)
}

func (n *node) getRune() (rune, bool) {
	// ELEMENT_BK are special zero-width elements that piggy-back onto other nodes.
	// If the node they're attached to has a non-zero-value blob (rune) then that should
	// also be rendered.
	// This means if an ELEMENT_BK is immediately followed by a null/zero byte,
	// that null/zero byte will not be rendered. This is not ideal, and the
	// whole ELEMENT_BK/timestamp subsystem needs an overhaul, but it's
	// probably fine in practice; timestamps should be at the start of lines,
	// null bytes shouldn't. And there's probably no useful/valid way to render
	// them anyway.
	if n.elem == nil || (n.elem.elementType == ELEMENT_BK && n.blob != 0) {
		return n.blob, true
	}
	return 0, false
}

// Whether the node has an element that persists when the node is overwritten.
// e.g. _bk;t=... timestamps are zero-width, placed in the _next_ node to be written.
// (That's _usually_ the start of a line, but not necessarily)
// The subsequent write to that next node needs to retain the element.
func (n *node) hasPersistentElement() bool {
	return n.elem != nil && n.elem.elementType == ELEMENT_BK
}
