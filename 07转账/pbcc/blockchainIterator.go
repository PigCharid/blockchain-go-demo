package pbcc

import (
	"log"
	"publicchain/conf"

	"github.com/boltdb/bolt"
)

//区块链迭代结构体
type BlockChainIterator struct {
	CurrentHash []byte   //当前区块的hash
	DB          *bolt.DB //数据库
}

//获取当前指向的区块，然后把指向改成上一个区块
func (bcIterator *BlockChainIterator) Next() *Block {
	block := new(Block)
	//打开数据库并读取
	err := bcIterator.DB.View(func(tx *bolt.Tx) error {
		//打开数据表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//根据当前hash获取数据并反序列化
			blockBytes := b.Get(bcIterator.CurrentHash)
			block = DeserializeBlock(blockBytes)
			//更新当前的hash
			bcIterator.CurrentHash = block.PrevBlockHash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block
}
