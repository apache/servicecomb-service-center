package tree

type Tree struct {
	Root *Node
}

func NewTree() *Tree {
	return new(Tree)
}

type Node struct {
	Res         interface{}
	left, right *Node
}

//binary sort tree
func (t *Tree) AddNode(n *Node, res interface{}, isAddToLeft func(addRes interface{}, node *Node) bool) *Node{
	if n == nil {
		n = new(Node)
		n.Res = res
		return n
	}
	if isAddToLeft(res, n) {
		n.left = t.AddNode(n.left, res, isAddToLeft)
	} else {
		n.right = t.AddNode(n.right, res, isAddToLeft)
	}
	return n
}

//middle oder traversal

func (t *Tree)MidOderTraversal(n *Node, handle func(res interface{}) error) error {
	if n == nil {
		return nil
	}

	var savedErr error

	err := t.MidOderTraversal(n.left, handle)
	if err != nil {
		savedErr = err
	}
	err = handle(n.Res)
	if err != nil {
		savedErr = err
	}
	err = t.MidOderTraversal(n.right, handle)
	if err != nil {
		savedErr = err
	}
	return savedErr
}

//todo add asynchronous handle handle func: go handle

//middle oder traversal
func (t *Tree)MidOderTraversalFailFast(n *Node, handle func(res interface{}) error) error {
	if n == nil {
		return nil
	}

	err := t.MidOderTraversalFailFast(n.left, handle)
	if err != nil {
		return err
	}
	err = handle(n.Res)
	if err != nil {
		return err
	}
	err = t.MidOderTraversalFailFast(n.right, handle)
	if err != nil {
		return err
	}
	return nil
}