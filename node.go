package terminal

var emptyNode = node{' ', &emptyStyle}

type node struct {
	blob  rune
	style *style
}

func (n *node) hasSameStyle(o node) bool {
	return n.style.isEqual(o.style)
}
