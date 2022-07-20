package pbcc

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//Block结构体
type Block struct {
	Height        int64  //高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	PrevBlockHash []byte //上一个区块的哈希值ProvHash：
	Data          []byte //交易数据Data：目前先设计为[]byte,后期是Transaction
	TimeStamp     int64  //时间戳TimeStamp：
	Hash          []byte //哈希值Hash：32个的字节，64个16进制数
	Nonce         int64  // 随机数
}

//创建新的区块
func NewBlock(data string, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, []byte(data), time.Now().Unix(), nil, 0}
	//调用工作量证明的方法，并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	// 然后把计算出来的结果赋给区块
	block.Hash = hash
	block.Nonce = nonce
	return block
}

//创建创世区块：
func CreateGenesisBlock(data string) *Block {
	return NewBlock(data, make([]byte, 32), 0)
}

//将区块序列化，得到一个字节数组---区块的行为
func (block *Block) Serilalize() []byte {
	//创建一个buffer
	var result bytes.Buffer
	//创建一个编码器
	encoder := gob.NewEncoder(&result)
	//编码--->打包
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
	//创建一个解码器
	decoder := gob.NewDecoder(reader)
	//解包
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
