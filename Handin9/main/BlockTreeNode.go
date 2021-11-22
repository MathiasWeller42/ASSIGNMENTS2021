package main

import (
	"fmt"
	"strconv"
	"strings"
)

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

//TODO FIX DO SHIT NOW Something is wrong when we try to add to a block that is not genesis - child goes missing... AMBER ALERT!!!

func MakeBlockTreeNode(vk string, slot int, draw string, blockData Block, signature string) *BlockTreeNode {
	blockTreeNode := new(BlockTreeNode)

	if slot == 0 {
		blockTreeNode.OwnBlockHash = "genesis"
		fmt.Println("Nice, I'm the genesis block!")
	} else {
		stringToHash := "BLOCK" + ":" + vk + ":" + strconv.Itoa(slot) + ":" + draw + ":" + strings.Join(blockData, ":") + ":" + signature
		blockTreeNode.OwnBlockHash = ConvertBigIntToString(Hash(stringToHash))
	}

	blockTreeNode.Block = "BLOCK"
	blockTreeNode.VK = vk
	blockTreeNode.Slot = slot
	blockTreeNode.Draw = draw
	blockTreeNode.BlockData = blockData
	blockTreeNode.TreeNodeSig = signature
	return blockTreeNode
}
