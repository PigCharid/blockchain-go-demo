package server

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"publicchain/conf"
	"publicchain/pbcc"

	"github.com/boltdb/bolt"
)

// 处理版本消息
func handleVersion(request []byte, bc *pbcc.BlockChain) {

	var buff bytes.Buffer
	var payload Version
	// 从请求中截出数据
	dataBytes := request[conf.COMMANDLENGTH:]

	// 反序列化 解析请求数据中的version消息到payload
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// 获取本节点存的链的区块高度
	bestHeight := bc.GetBestHeight()
	// 节点请求发来消息的区块高度
	foreignerBestHeight := payload.BestHeight
	// 如果本节点的区块高度大于发来消息节点的区块高度
	if bestHeight > foreignerBestHeight {
		//把本节点的区块链高度信息发给对方
		SendVersion(payload.AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		// 去向对方节点获取区块
		SendGetBlocks(payload.AddrFrom)
	}
	// 如果该节点之前没来同步过，那么加入已知节点的列表
	if !nodeIsKnown(payload.AddrFrom) {
		KnowNodes = append(KnowNodes, payload.AddrFrom)
	}

}

// 处理GetBlock消息
func handleGetblocks(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//获取所有区块的hash
	blocks := bc.GetBlockHashes()
	//向请求地址发送Inv消息
	SendInv(payload.AddrFrom, conf.BLOCK_TYPE, blocks)
}

// 处理Inv消息
func handleInv(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload Inv
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// 如果Inv消息的数据是Block类型
	if payload.Type == conf.BLOCK_TYPE {
		// 记录最新的区块hash
		blockHash := payload.Items[0]
		// 发送GetDate消息
		SendGetData(payload.AddrFrom, conf.BLOCK_TYPE, blockHash)
		// 如果携带的区块或者交易数量大于1
		if len(payload.Items) >= 1 {
			//存下其他剩余区块的hash
			TransactionArray = payload.Items[1:]
		}
	}
	// 如果Inv消息的数据是Tx类型
	if payload.Type == conf.TX_TYPE {
		// 获取最后一笔交易
		txHash := payload.Items[0]
		// 如果缓冲交易池里面没有这个交易，则像节点发送GetData数据
		if MemoryTxPool[hex.EncodeToString(txHash)] == nil {
			SendGetData(payload.AddrFrom, conf.TX_TYPE, txHash)
		}
	}
}

// 处理GetData消息
func handleGetData(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload GetData
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == conf.BLOCK_TYPE {
		// 获取区块消息
		block, err := bc.GetBlock([]byte(payload.Hash))
		if err != nil {
			return
		}
		SendBlock(payload.AddrFrom, block)
	}

	if payload.Type == conf.TX_TYPE {
		tx := MemoryTxPool[hex.EncodeToString(payload.Hash)]
		SendTx(payload.AddrFrom, tx)
	}
}

// 处理发送区块消息
func handleBlock(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload BlockData
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockBytes := payload.Block
	// 解析获取区块
	block := pbcc.DeserializeBlock(blockBytes)
	fmt.Println("Recevied a new block!")
	// 新的区块加入链上
	bc.AddBlock(block)
	fmt.Printf("Added block %x\n", block.Hash)
	// 如果还有区块
	if len(TransactionArray) > 0 {
		blockHash := TransactionArray[0]
		// 再去请求
		SendGetData(payload.AddrFrom, "block", blockHash)
		// 更新未打包进区块链的区块池
		TransactionArray = TransactionArray[1:]
	} else {
		fmt.Println("已经没有要处理的区块了")
	}
}

// 处理发送交易消息
func handleTx(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload Tx
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	tx := payload.Tx
	// 交易存到交易缓冲池子
	MemoryTxPool[hex.EncodeToString(tx.TxID)] = tx
	// 说明主节点自己
	if NodeAddress == KnowNodes[0] {
		// 给矿工节点发送交易hash
		for _, nodeAddr := range KnowNodes {
			if nodeAddr != NodeAddress && nodeAddr != payload.AddrFrom {
				SendInv("localhost:8002", conf.TX_TYPE, [][]byte{tx.TxID})
			}
		}
	}
	// 矿工进行挖矿验证
	if len(MemoryTxPool) >= 1 && len(MinerAddress) > 0 {
	MineTransactions:
		utxoSet := &pbcc.UTXOSet{bc}
		txs := []*pbcc.Transaction{tx}
		//奖励
		coinbaseTx := pbcc.NewCoinBaseTransaction(MinerAddress)
		txs = append(txs, coinbaseTx)
		_txs := []*pbcc.Transaction{}
		for _, tx := range txs {
			// 数字签名失败
			if bc.VerifyTransaction(tx, _txs) != true {
				log.Panic("ERROR: Invalid transaction")
			}
			_txs = append(_txs, tx)
		}
		//通过相关算法建立Transaction数组
		var block *pbcc.Block
		bc.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
			if b != nil {
				hash := b.Get([]byte("l"))
				blockBytes := b.Get(hash)
				block = pbcc.DeserializeBlock(blockBytes)
			}
			return nil
		})
		//建立新的区块
		block = pbcc.NewBlock(txs, block.Hash, block.Height+1)
		//将新区块存储到数据库
		bc.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
			if b != nil {
				b.Put(block.Hash, block.Serilalize())
				b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
			return nil
		})
		utxoSet.Update()
		SendBlock(KnowNodes[0], block.Serilalize())
		for _, tx := range txs {
			txID := hex.EncodeToString(tx.TxID)
			delete(MemoryTxPool, txID)
		}
		for _, node := range KnowNodes {
			if node != NodeAddress {
				SendInv(node, "block", [][]byte{block.Hash})
			}
		}
		if len(MemoryTxPool) > 0 {
			goto MineTransactions
		}
	}
}
