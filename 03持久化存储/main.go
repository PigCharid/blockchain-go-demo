package main

import "publicchain/pbcc"

func main() {
	blockchain := pbcc.CreateBlockChainWithGenesisBlock("i am genesis block")
	defer blockchain.DB.Close()

	//添加一个新区快
	blockchain.AddBlockToBlockChain("first Block")
	blockchain.AddBlockToBlockChain("second Block")
	blockchain.AddBlockToBlockChain("third Block")

	blockchain.PrintChains()
}
