package terminal

var emptyNode = node{' ', &emptyStyle}

type node struct {
	blob  rune
	style *style
}

func (n *node) hasSameStyle(o node) bool {
	return n == &o || n.style.string() == o.style.string()
}
