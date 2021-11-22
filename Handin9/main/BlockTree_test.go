package main

import (
	"fmt"
	"testing"
)

func TestGetLongestChain(t *testing.T) {
	tree1, tree2, tree3 := InitTreeNodes()
	ownBlockHash2 := tree2.Node.OwnBlockHash

	tree1.AddChild(tree2)
	tree1.AddChildAt(tree3, ownBlockHash2)
	blockChain := tree1.GetLongestChainOfBlocksAsSlice()
	fmt.Println("Here is the chain: ", blockChain)
	if len(blockChain) != 2 {
		fmt.Println("chain had the wrong length: ", len(blockChain))
	} else {
		fmt.Println("TestGetLongestChain passed")
	}

}

func TestSearch(t *testing.T) {
	tree1, tree2, _ := InitTreeNodes()
	ownBlockHash1 := tree1.Node.OwnBlockHash
	ownBlockHash2 := tree2.Node.OwnBlockHash

	fmt.Println("Searching for hash1: ")
	tree4 := tree1.Search(ownBlockHash1)
	tree4.PrintTree()

	fmt.Println("Searching for hash2: ")
	tree1.AddChild(tree2)
	tree5 := tree1.Search(ownBlockHash2)
	tree5.PrintTree()
}

func TestAddChildAt(t *testing.T) {
	tree1, tree2, tree3 := InitTreeNodes()
	//ownBlockHash1 := treeNode1.Node.OwnBlockHash
	ownBlockHash2 := tree2.Node.OwnBlockHash
	//ownBlockHash3 := treeNode3.Node.OwnBlockHash

	tree1.AddChild(tree2)
	fmt.Println("Now adding child at hash2:")
	tree1.AddChildAt(tree3, ownBlockHash2)
	tree2.AddChild(tree3)

	fmt.Println("")
	fmt.Println("tree1's printtree:")
	tree1.PrintTree()

	fmt.Println("tree2's children amount:", len(tree2.children))
	fmt.Println("")
	fmt.Println("Printtree tree2:")

	tree2.PrintTree()
}

func InitTreeNodes() (*BlockTree, *BlockTree, *BlockTree) {
	treeNode1 := MakeBlockTreeNode("vk1", 0, "draw1", nil, "signature1")
	treeNode2 := MakeBlockTreeNode("vk2", 1, "draw2", nil, "signature2")
	treeNode3 := MakeBlockTreeNode("vk3", 2, "draw3", nil, "signature3")
	return MakeBlockTree(treeNode1), MakeBlockTree(treeNode2), MakeBlockTree(treeNode3)

}
