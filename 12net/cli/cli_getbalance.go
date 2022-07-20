package cli

import (
	"fmt"
	"os"
	"publicchain/pbcc"
)

//查询余额
func (cli *CLI) getBalance(address string, nodeID string) {
	fmt.Println("查询余额：", address)
	bc := pbcc.GetBlockchainObject(nodeID)
	if bc == nil {
		fmt.Println("数据库不存在，无法查询")
		os.Exit(1)
	}
	defer bc.DB.Close()
	utxoSet := &pbcc.UTXOSet{bc}
	utxoSet.ResetUTXOSet()
	balance := utxoSet.GetBalance(address)
	fmt.Printf("%s,一共有%d个Token\n", address, balance)
}
