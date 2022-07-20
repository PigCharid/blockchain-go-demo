package pbcc

import (
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
