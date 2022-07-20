package cli

import "publicchain/pbcc"

func (cli *CLI) createGenesisBlockchain(data string) {
	pbcc.CreateBlockChainWithGenesisBlock(data)
}
