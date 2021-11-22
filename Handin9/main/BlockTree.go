package main

import "fmt"

type BlockTree struct {
	Node     *BlockTreeNode
	children []*BlockTree
	parent   *BlockTree
}

func MakeBlockTree(root *BlockTreeNode) *BlockTree {
	blockTree := new(BlockTree)
	blockTree.Node = root
	blockTree.parent = nil
	blockTree.children = make([]*BlockTree, 0)

	return blockTree
}

func (blockTree *BlockTree) AddChild(tree *BlockTree) {
	blockTree.children = append(blockTree.children, tree)
	tree.parent = blockTree
}

func (blockTree *BlockTree) AddChildAt(tree *BlockTree, blockHash string) {
	foundTree := blockTree.Search(blockHash)
	if foundTree == nil {
		fmt.Println("Tried to find a node by a hash that does not exist:", blockHash)
	} else {
		fmt.Println("Adding to block with slotNumber:", foundTree.Node.Slot)
		foundTree.AddChild(tree)
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
	if blockTree.Node.OwnBlockHash == "genesis" {
		fmt.Println("The genesis node has", len(blockTree.children), "children")
	} else {
		fmt.Println("This node with slot", blockTree.Node.Slot, "has", len(blockTree.children), "children and parent slot", blockTree.parent.Node.Slot)
	}
	for _, blockTreeChild := range blockTree.children {
		blockTreeChild.PrintTree()
	}
}

func (blockTree *BlockTree) GetLongestChainOfBlocksAsSlice() []Block {
	leaf := blockTree.GetLongestChainLeaf()
	blockList := make([]Block, 0)
	currentTree := leaf
	for {
		if currentTree.Node.OwnBlockHash == "genesis" {
			return blockList
		} else {
			block := currentTree.Node.BlockData
			block = append(block, currentTree.Node.VK)
			blockList = append([]Block{block}, blockList...)
			currentTree = currentTree.parent
		}
	}
}

func (blockTree *BlockTree) GetTreeSize() int {
	amount := 0
	if blockTree.Node.OwnBlockHash == "genesis" {
		amount += 1
	}

	for _, blockTreeChild := range blockTree.children {
		amount += blockTreeChild.GetTreeSize()
	}
	return amount + len(blockTree.children)
}
