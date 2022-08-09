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
	if n.elem != nil {
		return 0, false
	}
	return n.blob, true
}
