package pbcc

import (
	"bytes"
	"crypto/sha256"
	"publicchain/utils"
	"strconv"
	"time"
)

//Block结构体
type Block struct {
	Height        int64  //高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	PrevBlockHash []byte //上一个区块的哈希值ProvHash：
	Data          []byte //交易数据Data：目前先设计为[]byte,后期是Transaction
	TimeStamp     int64  //时间戳TimeStamp：
	Hash          []byte //哈希值Hash：32个的字节，64个16进制数
}

//设置区块的hash
func (block *Block) SetHash() {
	//将高度转为字节数组
	heightBytes := utils.IntToHex(block.Height)
	// 先将时间戳按二进制转化成字符串
	timeString := strconv.FormatInt(block.TimeStamp, 2)
	// 强转一下
	timeBytes := []byte(timeString)
	//拼接所有的属性 把几个属性按照下面的空字节来分割拼接
	blockBytes := bytes.Join([][]byte{
		heightBytes,
		block.PrevBlockHash,
		block.Data,
		timeBytes}, []byte{})
	//生成哈希值，返回一个32位的字节数组 256位
	//fmt.Println(blockBytes)
	hash := sha256.Sum256(blockBytes)
	block.Hash = hash[:]
}

//创建新的区块
func NewBlock(data string, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, []byte(data), time.Now().Unix(), nil}
	//设置哈希值
	block.SetHash()
	return block
}

//创建创世区块：
func CreateGenesisBlock(data string) *Block {
	return NewBlock(data, make([]byte, 32), 0)
}
