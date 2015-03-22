package terminal

var emptyNode = node{blob: ' ', style: &emptyStyle}

type node struct {
	blob  rune
	style *style
	image *itermImage
}

func (n *node) hasSameStyle(o node) bool {
	return n.style.isEqual(o.style)
}
