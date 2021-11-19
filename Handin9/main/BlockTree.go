package main

type BlockTree struct {
	Node     BlockTreeNode
	children []BlockTree
}

func MakeBlockTree(root BlockTreeNode) *BlockTree {
	blockTree := new(BlockTree)
	blockTree.Node = root
	blockTree.children = make([]BlockTree, 0)
	return blockTree
}

func (blockTree *BlockTree) AddChild(node BlockTreeNode) {
	blockTree.children = append(blockTree.children, *MakeBlockTree(node))
}

func (blockTree *BlockTree) AddChildAt(node BlockTreeNode, blockHash string) {
	found := blockTree.Search(blockHash)
	found.children = append(found.children, *MakeBlockTree(node))
}

func (blockTree *BlockTree) Search(blockHash string) *BlockTree {
	for _, blockTree := range blockTree.children {
		if blockTree.Node.OwnBlockHash == blockHash {
			return &blockTree
		}
		foundTree := blockTree.Search(blockHash)
		if foundTree != nil {
			return foundTree
		}
	}
	return nil
}
