package main

type BlockTree struct {
	Node     BlockTreeNode
	children []BlockTree
}

func MakeNewBlockTree(root BlockTreeNode) *BlockTree {
	blockTree := new(BlockTree)
	blockTree.Node = root
	blockTree.children = make([]BlockTree, 0)
	return blockTree
}

func (blockTree *BlockTree) AppendToTree(node BlockTreeNode) *BlockTree {

	return MakeNewBlockTree(node)
}
