package pbcc

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//Block结构体
type Block struct {
	//字段：
	//高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	Height int64
	//上一个区块的哈希值ProvHash：
	PrevBlockHash []byte
	//交易数据Data：目前先设计为[]byte,后期是Transaction
	//Data []byte
	Txs []*Transaction
	//时间戳TimeStamp：
	TimeStamp int64
	//哈希值Hash：32个的字节，64个16进制数
	Hash []byte
	// 随机数
	Nonce int64
}

//创建新的区块
func NewBlock(txs []*Transaction, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, txs, time.Now().Unix(), nil, 0}
	//调用工作量证明的方法，并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}

//创建创世区块：
func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, make([]byte, 32), 0)
}

//将区块序列化，得到一个字节数组---区块的行为
func (block *Block) Serilalize() []byte {
	//1.创建一个buffer
	var result bytes.Buffer
	//2.创建一个编码器
	encoder := gob.NewEncoder(&result)
	//3.编码--->打包
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化，得到一个区块
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	var reader = bytes.NewReader(blockBytes)
	//1.创建一个解码器
	decoder := gob.NewDecoder(reader)
	//解包
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

//将Txs转为[]byte
func (block *Block) HashTransactions() []byte {
	var txs [][]byte
	for _, tx := range block.Txs {
		txs = append(txs, tx.Serialize())
	}
	mTree := NewMerkleTree(txs)
	return mTree.RootNode.Data
}
