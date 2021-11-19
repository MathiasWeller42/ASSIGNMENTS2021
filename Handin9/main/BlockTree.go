package main

import "fmt"

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
	if found == nil {
		fmt.Println("Tried to find a node by a hash that does not exist:", blockHash)
	} else {
		fmt.Println("Added a child by looking for a hash")
		found.AddChild(node)
	}
}

func (blockTree *BlockTree) Search(blockHash string) *BlockTree {
	if blockTree.Node.OwnBlockHash == blockHash {
		return blockTree
	}
	for _, blockTreeChild := range blockTree.children {
		foundTree := blockTreeChild.Search(blockHash)
		if foundTree != nil {
			return foundTree
		}
	}
	return nil
}

func (blockTree *BlockTree) GetLongestChainLeaf() *BlockTree {
	foundBlockTree, _ := blockTree.getPointerToPrevBlockAux(1)
	return foundBlockTree
}

func (blockTree *BlockTree) getPointerToPrevBlockAux(currentLength int) (*BlockTree, int) {
	maxLength := 0
	var maxBlockPointer *BlockTree
	if len(blockTree.children) == 0 {
		return blockTree, currentLength
	}
	for _, blockTreeChild := range blockTree.children {
		foundBlockPointer, foundLength := blockTreeChild.getPointerToPrevBlockAux(currentLength + 1)
		if foundLength > maxLength {
			maxLength = foundLength
			maxBlockPointer = foundBlockPointer
		}
	}
	return maxBlockPointer, maxLength
}

func (blockTree *BlockTree) PrintTree() {
	fmt.Println("This node with slot", blockTree.Node.Slot, "has", len(blockTree.children), "children")
	for _, blockTreeChild := range blockTree.children {
		blockTreeChild.PrintTree()
	}
}
