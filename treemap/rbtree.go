package treemap

type color bool

const (
	black, red color = true, false
)

type Tree[K any, V any] struct {
	Root       *Node[K, V]
	size       int
	Comparator Comparator[K]
}

func (tree *Tree[K, V]) rotateLeft(node *Node[K, V]) {

}

func (tree *Tree[K, V]) rotateRight(node *Node[K, V]) {

}

func (tree *Tree[K, V]) replaceNode(old *Node[K, V], new *Node[K, V]) {
	if old.Parent == nil {
		tree.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

func (tree *Tree[K, V]) rebalance(node *Node[K, V]) {

}

type Node[K any, V any] struct {
	Key    K
	Value  V
	color  color
	Left   *Node[K, V]
	Right  *Node[K, V]
	Parent *Node[K, V]
}

func (node *Node[K, V]) uncle() *Node[K, V] {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *Node[K, V]) sibling() *Node[K, V] {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

func (node *Node[K, V]) grandparent() *Node[K, V] {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.Parent
}

func NewWithComparator[K any, V any](comparator Comparator[K]) *Tree[K, V] {
	return &Tree[K, V]{Comparator: comparator}
}

func (tree *Tree[K, V]) Put(key K, value V) {
	var insertedNode *Node[K, V]
	if tree.Root == nil {
		tree.Comparator(key, key)
		tree.Root = &Node[K, V]{Key: key, Value: value, color: red}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {
			compare := tree.Comparator(key, node.Key)
			switch {
			case compare == 0:
				node.Key = key
				node.Value = value
				return
			case compare < 0:
				if node.Left == nil {
					node.Left = &Node[K, V]{Key: key, Value: value, Parent: node, color: red}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case compare > 0:
				if node.Right == nil {
					node.Right = &Node[K, V]{Key: key, Value: value, Parent: node, color: red}
					insertedNode = node.Right
					loop = false
				}
			}
		}
		insertedNode.Parent = node
	}
	// tree.insertCase1(insertedNode)
	tree.size++
}

func (tree *Tree[K, V]) insertCase1(node *Node[K, V]) {
	if node.Parent == nil {
		node.color = black
	} else {
		tree.insertCase2(node)
	}
}

func (tree *Tree[K, V]) insertCase2(node *Node[K, V]) {
	if nodeColor(node) == black {
		return
	}
	tree.insertCase3(node)
}

func (tree *Tree[K, V]) insertCase3(node *Node[K, V]) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.Parent.color = black
		uncle.color = black
		node.grandparent().color = red
		tree.insertCase1(node.grandparent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree[K, V]) insertCase4(node *Node[K, V]) {
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		tree.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		tree.rotateRight(node.Parent)
		node = node.Right
	}
	tree.insertCase5(node)
}

func (tree *Tree[K, V]) insertCase5(node *Node[K, V]) {
	node.Parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red

}

func nodeColor[K any, V any](node *Node[K, V]) color {
	if node == nil {
		return black
	}
	return node.color
}
