package cli

import (
	"fmt"
	"publicchain/pbcc"
)

// 输出UXTO
func (cli *CLI) TestMethod(nodeID string) {
	blockchain := pbcc.GetBlockchainObject(nodeID)
	unSpentOutputMap := blockchain.FindUnSpentOutputMap()
	fmt.Println(unSpentOutputMap)
	for key, value := range unSpentOutputMap {
		fmt.Println(key)
		for _, utxo := range value.UTXOS {
			fmt.Println("金额：", utxo.Output.Value)
			fmt.Printf("地址：%v\n", utxo.Output.PubKeyHash)
			fmt.Println("---------------------")
		}
	}
	utxoSet := &pbcc.UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}
