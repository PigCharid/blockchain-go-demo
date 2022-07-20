package pbcc

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"publicchain/conf"
	"publicchain/utils"
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
	//1.获取迭代器对象
	bcIterator := bc.Iterator()

	//2.循环迭代
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
				fmt.Printf("\t\t\tScriptSiq:%s\n", in.ScriptSiq)
			}
			fmt.Println("\t\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\t\tvalue:%d\n", out.Value)
				fmt.Printf("\t\t\tScriptPubKey:%s\n", out.ScriptPubKey)
			}
		}
		fmt.Printf("\t时间:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数/;%d\n", block.Nonce)
		//直到父hash值为0
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
	var unUTXOs []*UTXO                      //未花费输出
	spentTxOutputs := make(map[string][]int) //存储已经花费

	//添加先从txs遍历，查找未花费
	for i := len(txs) - 1; i >= 0; i-- {
		// 计算某笔交易中的输出是否已经被花费了
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
	//判断是否为铸币交易，只有费铸币交易才有输入，先记录这个交易里面该地址的输入
	if !tx.IsCoinbaseTransaction() {
		//遍历交易里面的输入
		for _, in := range tx.Vins {
			//如果解锁 也就是这个输出是该地址花掉的了
			if in.UnLockWithAddress(address) {
				//回去交易的Hash
				key := hex.EncodeToString(in.TxID)
				// 记录已经花费的  记录交易的hash和索引
				spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
			}
		}
	}
outputs:
	// 遍历交易的输出
	for index, out := range tx.Vouts {
		// 如果解锁 那么这个输出是属于该地址
		if out.UnLockWithAddress(address) {
			//说明交易中有我的输出花费 需要进行对比
			if len(spentTxOutputs) != 0 {
				var isSpentUTXO bool
				// 遍历已经花费的输入
				for txID, indexArray := range spentTxOutputs {
					for _, i := range indexArray {
						//下标对上 交易的ID对上
						if i == index && txID == hex.EncodeToString(tx.TxID) {
							// 说明这个输出是被花掉的
							isSpentUTXO = true
							continue outputs //那么继续下一个
						}
					}
				}
				// 如果没花掉，那么加入utxo
				if !isSpentUTXO {
					// 组装一个UTXO对象
					utxo := &UTXO{tx.TxID, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}

			} else { //也就是该地址没有花费  则把属于该地址的输出都纳入UTXO
				utxo := &UTXO{tx.TxID, index, out}
				unUTXOs = append(unUTXOs, utxo)
			}
		}
	}
	return unUTXOs
}
