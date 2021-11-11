package main

type BlockTreeNode struct {
	Block         string
	VK            string
	Slot          int
	Draw          string
	BlockData     Block
	PrevBlockHash string
	TreeNodeSig   string
}

func MakeBlockTreeNode(vk string, slot int, draw string, block Block, prevBlockHash string, signature string) *BlockTreeNode {
	blockTreeNode := new(BlockTreeNode)
	blockTreeNode.Block = "BLOCK"
	blockTreeNode.VK = vk
	blockTreeNode.Slot = slot
	blockTreeNode.Draw = draw
	blockTreeNode.BlockData = block
	blockTreeNode.PrevBlockHash = prevBlockHash
	blockTreeNode.TreeNodeSig = signature
	return blockTreeNode
}
