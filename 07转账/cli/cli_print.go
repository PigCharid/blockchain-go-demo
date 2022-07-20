package cli

import (
	"fmt"
	"os"
	"publicchain/pbcc"
)

func (cli *CLI) printChains() {
	bc := pbcc.GetBlockchainObject()
	if bc == nil {
		fmt.Println("未创建，没有区块可以打印")
		os.Exit(1)
	}
	defer bc.DB.Close()
	bc.PrintChains()
}
