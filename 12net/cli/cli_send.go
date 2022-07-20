package cli

import (
	"fmt"
	"publicchain/pbcc"
	"publicchain/server"
	"strconv"
)

//转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {
	blockchain := pbcc.GetBlockchainObject(nodeID)
	utxoSet := &pbcc.UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
	defer blockchain.DB.Close()
	if mineNow {
		blockchain.MineNewBlock(from, to, amount, nodeID)
		//转账成功以后，需要更新一下
		utxoSet.Update()
	} else {
		// 把交易发送到矿工节点去进行验证
		fmt.Println("由矿工节点处理......")
		value, _ := strconv.Atoi(amount[0])
		tx := pbcc.NewSimpleTransaction(from[0], to[0], int64(value), utxoSet, []*pbcc.Transaction{}, nodeID)
		// 向全节点发送一下
		server.SendTx(server.KnowNodes[0], tx)
	}
}
