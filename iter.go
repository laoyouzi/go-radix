package radix

import (
	"sort"
	"strings"
)

// getLowerBoundEdge returns the edge with label >= the given label.
func (n *node) getLowerBoundEdge(label byte) (int, *node) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num {
		return idx, n.edges[idx].node
	}
	return -1, nil
}

// Iterator walks tree keys in lexicographic order without scanning from the root
// on every step. Ported from hashicorp/go-immutable-radix (MPL-2.0).
type Iterator struct {
	root  *node
	node  *node
	stack []edges
}

// Iterator returns an iterator over the whole tree.
func (t *Tree) Iterator() *Iterator {
	if t == nil {
		return &Iterator{}
	}
	return &Iterator{root: t.root, node: t.root}
}

// SeekPrefix positions the iterator at the first key with the given prefix.
func (i *Iterator) SeekPrefix(prefix string) {
	if i == nil || i.root == nil {
		i.node = nil
		i.stack = nil
		return
	}
	i.stack = nil
	n := i.root
	search := prefix
	for {
		if len(search) == 0 {
			i.node = n
			return
		}

		n = n.getEdge(search[0])
		if n == nil {
			i.node = nil
			return
		}

		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else if strings.HasPrefix(n.prefix, search) {
			i.node = n
			return
		} else {
			i.node = nil
			return
		}
	}
}

func (i *Iterator) recurseMin(n *node) *node {
	if n.leaf != nil {
		return n
	}
	nEdges := len(n.edges)
	if nEdges > 1 {
		i.stack = append(i.stack, n.edges[1:])
	}
	if nEdges > 0 {
		return i.recurseMin(n.edges[0].node)
	}
	return nil
}

// SeekLowerBound positions the iterator at the smallest key >= key.
func (i *Iterator) SeekLowerBound(key string) {
	if i == nil || i.root == nil {
		i.node = nil
		i.stack = nil
		return
	}
	i.stack = []edges{}
	n := i.root
	i.node = nil
	search := key

	found := func(n *node) {
		i.stack = append(i.stack, edges{edge{node: n}})
	}

	findMin := func(n *node) {
		n = i.recurseMin(n)
		if n != nil {
			found(n)
		}
	}

	for n != nil {
		var prefixCmp int
		if len(n.prefix) < len(search) {
			prefixCmp = strings.Compare(n.prefix, search[:len(n.prefix)])
		} else {
			prefixCmp = strings.Compare(n.prefix, search)
		}

		if prefixCmp > 0 {
			findMin(n)
			return
		}

		if prefixCmp < 0 {
			i.node = nil
			return
		}

		if n.leaf != nil && n.leaf.key == key {
			found(n)
			return
		}

		search = search[len(n.prefix):]

		if len(search) == 0 {
			findMin(n)
			return
		}

		idx, lbNode := n.getLowerBoundEdge(search[0])
		if lbNode == nil {
			return
		}

		if idx+1 < len(n.edges) {
			i.stack = append(i.stack, n.edges[idx+1:])
		}

		n = lbNode
	}
}

// Next returns the next key/value in order.
func (i *Iterator) Next() (string, interface{}, bool) {
	if i.stack == nil && i.node != nil {
		i.stack = []edges{{edge{node: i.node}}}
	}

	for len(i.stack) > 0 {
		n := len(i.stack)
		last := i.stack[n-1]
		elem := last[0].node

		if len(last) > 1 {
			i.stack[n-1] = last[1:]
		} else {
			i.stack = i.stack[:n-1]
		}

		if len(elem.edges) > 0 {
			i.stack = append(i.stack, elem.edges)
		}

		if elem.leaf != nil {
			return elem.leaf.key, elem.leaf.val, true
		}
	}
	return "", nil, false
}
