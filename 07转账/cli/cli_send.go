package cli

import (
	"fmt"
	"os"
	"publicchain/pbcc"
)

//转账
func (cli *CLI) send(from, to, amount []string) {
	blockchain := pbcc.GetBlockchainObject()
	if blockchain == nil {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}
	blockchain.MineNewBlock(from, to, amount)
	defer blockchain.DB.Close()
}
