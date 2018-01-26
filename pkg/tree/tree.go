package tree

//The tree is binary sort tree
type tree struct {
	root *Node
	isAddToLeft func(node *Node, addRes interface{}) bool
}

func NewTree(isAddToLeft func(node *Node, addRes interface{}) bool) *tree {
	return &tree{
		isAddToLeft: isAddToLeft,
	}
}

type Node struct {
	Res         interface{}
	left, right *Node
}

func (t *tree) GetRoot()*Node {
	return t.root
}

//add res into tree
func (t *tree) AddNode(res interface{}) *Node {
	return t.addNode(t.root, res)
}

func (t *tree) addNode(n *Node, res interface{}) *Node{
	if n == nil {
		n = new(Node)
		n.Res = res
		if t.root == nil {
			t.root = n
		}
		return n
	}
	if t.isAddToLeft(n, res) {
		n.left = t.addNode(n.left, res)
	} else {
		n.right = t.addNode(n.right, res)
	}
	return n
}

//middle oder traversal, handle is the func that deals with the res, n is the start node to traversal
func (t *tree)MidOderTraversal(n *Node, handle func(res interface{}) error) error {
	if n == nil {
		return nil
	}

	err := t.MidOderTraversal(n.left, handle)
	if err != nil {
		return err
	}
	err = handle(n.Res)
	if err != nil {
		return err
	}
	err = t.MidOderTraversal(n.right, handle)
	if err != nil {
		return err
	}
	return nil
}

//todo add asynchronous handle handle func: go handle