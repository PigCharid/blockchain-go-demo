package pbcc

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"publicchain/conf"
	"publicchain/crypto"
	"publicchain/utils"
	"publicchain/wallet"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

//创建区块链
type BlockChain struct {
	Tip []byte   // 最新区块的Hash值
	DB  *bolt.DB //数据库对象
}

//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(address string) {
	if utils.DBExists() {
		fmt.Println("数据库已经存在")
		return
	}
	fmt.Println("数据库不存在,创建创世区块：")
	//先创建coinbase交易
	txCoinBase := NewCoinBaseTransaction(address)
	// 创世区块
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})
	//打开数据库
	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(conf.BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误")
			}
			//存储最新区块的hash
			b.Put([]byte("l"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := &BlockChain{genesisBlock.Hash, db}
	utxoSet := &UTXOSet{bc}
	utxoSet.ResetUTXOSet()
}

//添加一个新的区块，到区块链中
func (bc *BlockChain) AddBlockToBlockChain(txs []*Transaction) {
	//更新数据库
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//根据最新块的hash读取数据，并反序列化最后一个区块
			blockBytes := b.Get(bc.Tip)
			lastBlock := DeserializeBlock(blockBytes)
			//创建新的区块
			newBlock := NewBlock(txs, lastBlock.Hash, lastBlock.Height+1)
			//将新的区块序列化并存储
			err := b.Put(newBlock.Hash, newBlock.Serilalize())
			if err != nil {
				log.Panic(err)
			}
			//更新最后一个哈希值，以及blockchain的tip
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//获取一个迭代器
func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.Tip, bc.DB}
}

// 借助迭代器输出区块链
func (bc *BlockChain) PrintChains() {
	//获取迭代器对象
	bcIterator := bc.Iterator()

	//循环迭代
	for {
		block := bcIterator.Next()
		fmt.Printf("第%d个区块的信息:\n", block.Height+1)
		//获取当前hash对应的数据，并进行反序列化
		fmt.Printf("\t高度:%d\n", block.Height)
		fmt.Printf("\t上一个区块的hash:%x\n", block.PrevBlockHash)
		fmt.Printf("\t当前的hash:%x\n", block.Hash)
		//fmt.Printf("\t数据：%v\n", block.Txs)
		fmt.Println("\t交易:")
		for _, tx := range block.Txs {
			fmt.Printf("\t\t交易ID:%x\n", tx.TxID)
			fmt.Println("\t\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\t\tTxID:%x\n", in.TxID)
				fmt.Printf("\t\t\tVout:%d\n", in.Vout)
				fmt.Printf("\t\t\tPublicKey:%v\n", in.PublicKey)
			}
			fmt.Println("\t\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\t\tvalue:%d\n", out.Value)
				fmt.Printf("\t\t\tPubKeyHash:%v\n", out.PubKeyHash)
			}
		}
		fmt.Printf("\t时间:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数:%d\n", block.Nonce)

		//3.直到父hash值为0
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}

// 获取最新的区块链
func GetBlockchainObject() *BlockChain {
	if !utils.DBExists() {
		fmt.Println("数据库不存在，无法获取区块链")
		return nil
	}

	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var blockchain *BlockChain
	//读取数据库
	err = db.View(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//读取最后一个hash
			hash := b.Get([]byte("l"))
			//创建blockchain
			blockchain = &BlockChain{hash, db}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return blockchain
}

//找到某地址对应的所有UTXO
func (bc *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	/*
		1.先遍历未打包的交易(参数txs)，找出未花费的Output，为什么校验未打包的交易，教程有解释
		2.遍历数据库，获取每个块中的Transaction，找出未花费的Output。
	*/
	var unUTXOs []*UTXO                      //未花费
	spentTxOutputs := make(map[string][]int) //存储已经花费

	//添加先从txs遍历，查找未花费
	for i := len(txs) - 1; i >= 0; i-- {
		unUTXOs = caculate(txs[i], address, spentTxOutputs, unUTXOs)
	}

	bcIterator := bc.Iterator()
	for {
		block := bcIterator.Next()
		//统计未花费
		//获取block中的每个Transaction
		for i := len(block.Txs) - 1; i >= 0; i-- {
			unUTXOs = caculate(block.Txs[i], address, spentTxOutputs, unUTXOs)
		}

		//结束迭代
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
	return unUTXOs
}

//找出这比交易中的属于对应地址的UTXO
func caculate(tx *Transaction, address string, spentTxOutputs map[string][]int, unUTXOs []*UTXO) []*UTXO {
	//先遍历TxInputs，表示花费
	if !tx.IsCoinbaseTransaction() {
		for _, in := range tx.Vins {
			//如果解锁
			fullPayloadHash := crypto.Base58Decode([]byte(address))
			pubKeyHash := fullPayloadHash[1 : len(fullPayloadHash)-conf.AddressChecksumLen]

			if in.UnLockWithAddress(pubKeyHash) {
				key := hex.EncodeToString(in.TxID)
				spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
			}
		}
	}
outputs:
	for index, out := range tx.Vouts {
		if out.UnLockWithAddress(address) {
			if len(spentTxOutputs) != 0 {
				var isSpentUTXO bool

				for txID, indexArray := range spentTxOutputs {
					for _, i := range indexArray {
						if i == index && txID == hex.EncodeToString(tx.TxID) {
							isSpentUTXO = true
							continue outputs
						}
					}
				}
				if !isSpentUTXO {
					utxo := &UTXO{tx.TxID, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}

			} else {
				utxo := &UTXO{tx.TxID, index, out}
				unUTXOs = append(unUTXOs, utxo)
			}
		}
	}
	return unUTXOs
}

//查找UTXO的升级版
// 找出来以后判断余额够不够 然后把可花费的UTXO改成map[交易ID]输出的下标
func (bc *BlockChain) FindSpendableUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	var balance int64
	utxos := bc.UnUTXOs(from, txs)
	spendableUTXO := make(map[string][]int)
	for _, utxo := range utxos {
		balance += utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxID)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		if balance >= amount {
			break
		}
	}
	if balance < amount {
		fmt.Printf("%s 余额不足，总额：%d,需要：%d\n", from, balance, amount)
		os.Exit(1)
	}
	return balance, spendableUTXO
}

//挖掘新的区块 有交易的时候就会调用
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction
	//奖励
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)

	utxoSet := &UTXOSet{bc}

	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		tx := NewSimpleTransaction(from[i], to[i], amountInt, utxoSet, txs)
		txs = append(txs, tx)
	}

	var block *Block    //数据库中的最后一个block
	var newBlock *Block //要创建的新的block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) //数据库中的最后一个block
		}
		return nil
	})

	//在建立新区块钱，对txs进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx, _txs) != true {
			log.Panic("签名验证失败。。")
		}
		_txs = append(_txs, tx)
	}

	newBlock = NewBlock(txs, block.Hash, block.Height+1)

	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
}

// 获取余额
func (bc *BlockChain) GetBalance(address string, txs []*Transaction) int64 {
	unUTXOs := bc.UnUTXOs(address, txs)
	var amount int64
	for _, utxo := range unUTXOs {
		amount = amount + utxo.Output.Value
	}
	return amount
}

//根据交易ID查找对应的Transaction
func (bc *BlockChain) FindTransactionByTxID(txID []byte, txs []*Transaction) *Transaction {
	itertaor := bc.Iterator()
	//先遍历txs
	for _, tx := range txs {
		if bytes.Equal(txID, tx.TxID) {
			return tx
		}
	}

	for {
		block := itertaor.Next()
		for _, tx := range block.Txs {
			if bytes.Equal(txID, tx.TxID) {
				return tx
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return &Transaction{}
}

// 区块链层的交易签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTxs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTxs[hex.EncodeToString(prevTx.TxID)] = prevTx
	}

	tx.Sign(privKey, prevTxs)
}

//区块链层的交易签名验证
func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {
	prevTXs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTXs[hex.EncodeToString(prevTx.TxID)] = prevTx
	}
	return tx.Verify(prevTXs)
}

//查询未花费的Output map[string] *TxOutputs
func (bc *BlockChain) FindUnSpentOutputMap() map[string]*TxOutputs {
	iterator := bc.Iterator()

	//存储已经花费：·[txID], txInput
	spentUTXOsMap := make(map[string][]*TXInput)
	//存储未花费
	unSpentOutputMaps := make(map[string]*TxOutputs)
	for {
		block := iterator.Next()
		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TxOutputs{[]*UTXO{}}
			tx := block.Txs[i]
			if !tx.IsCoinbaseTransaction() {
				for _, txInput := range tx.Vins {
					key := hex.EncodeToString(txInput.TxID)
					spentUTXOsMap[key] = append(spentUTXOsMap[key], txInput)
				}
			}
			txID := hex.EncodeToString(tx.TxID)
		work:
			for index, out := range tx.Vouts {
				txInputs := spentUTXOsMap[txID]
				if len(txInputs) > 0 {
					var isSpent bool
					for _, input := range txInputs {
						inputPubKeyHash := wallet.PubKeyHash(input.PublicKey)
						if bytes.Equal(inputPubKeyHash, out.PubKeyHash) {
							if input.Vout == index {
								isSpent = true
								continue work
							}
						}
					}
					if !isSpent {
						utxo := &UTXO{tx.TxID, index, out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}

				} else {
					utxo := &UTXO{tx.TxID, index, out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			//设置
			unSpentOutputMaps[txID] = txOutputs
		}

		//停止迭代
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unSpentOutputMaps
}
