package terminal

var emptyNode = node{blob: ' '}

// node represents an item in the screen. Most of the time, it is a single rune
// which may or may not have a style (colour, bold, etc).
// Sometimes it is a HTML element (e.g. from an inline image). This is encoded
// by using style.element() == true and using blob as the index into a slice of
// elements stored in the line.
type node struct {
	blob  rune
	style style
}

// hasSameStyle reports if the two nodes have the same style.
func (n *node) hasSameStyle(o node) bool {
	return n.style&styleComparisonMask == o.style&styleComparisonMask
}
