package cli

import "publicchain/pbcc"

// 创建区块链
func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	pbcc.CreateBlockChainWithGenesisBlock(address, nodeID)
	bc := pbcc.GetBlockchainObject(nodeID)
	defer bc.DB.Close()
	if bc != nil {
		utxoSet := &pbcc.UTXOSet{bc}
		utxoSet.ResetUTXOSet()
	}
}
