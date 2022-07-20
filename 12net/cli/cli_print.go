package cli

import (
	"fmt"
	"os"
	"publicchain/pbcc"
)

// 打印节点的区块链信息
func (cli *CLI) printChains(nodeID string) {
	bc := pbcc.GetBlockchainObject(nodeID)
	if bc == nil {
		fmt.Println("没有区块可以打印")
		os.Exit(1)
	}
	defer bc.DB.Close()
	bc.PrintChains()
}
