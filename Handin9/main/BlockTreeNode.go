package main

type BlockTreeNode struct {
	Block         string
	VK            string
	Slot          int
	Draw          string
	BlockData     Block
	OwnBlockHash  string
	PrevBlockHash string
	TreeNodeSig   string
}

func MakeBlockTreeNode(vk string, slot int, draw string, blockData Block, ownBlockHash string, prevBlockHash string, signature string) *BlockTreeNode {
	blockTreeNode := new(BlockTreeNode)
	blockTreeNode.Block = "BLOCK"
	blockTreeNode.VK = vk
	blockTreeNode.Slot = slot
	blockTreeNode.Draw = draw
	blockTreeNode.BlockData = blockData
	blockTreeNode.OwnBlockHash = ownBlockHash
	blockTreeNode.PrevBlockHash = prevBlockHash
	blockTreeNode.TreeNodeSig = signature
	return blockTreeNode
}
