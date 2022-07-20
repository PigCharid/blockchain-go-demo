package main

import (
	"fmt"
	"publicchain/pbcc"
)

func main() {
	//创建带有创世区块的区块链
	blockchain := pbcc.CreateBlockChainWithGenesisBlock("i am genesisblock")
	//添加一个新区快
	blockchain.AddBlockToBlockChain("first Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("second Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("third Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	for _, block := range blockchain.Blocks {
		fmt.Printf("Timestamp: %d\n", block.TimeStamp)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("height: %d\n", block.Height)
		fmt.Println("--------------------------------------------")
	}
}
