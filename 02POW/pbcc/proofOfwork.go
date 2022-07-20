package pbcc

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"publicchain/conf"
	"publicchain/utils"
)

//pow结构体
type ProofOfWork struct {
	Block  *Block   //要验证的区块
	Target *big.Int //大整数存储,目标哈希
}

//创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {
	/**
	target计算方式  假设：Hash为8位，targetBit为2位
	eg:0000 0001(8位的Hash)
	1.8-2 = 6 将上值左移6位
	2.0000 0001 << 6 = 0100 0000 = target
	3.只要计算的Hash满足 ：hash < target，便是符合POW的哈希值
	*/

	//创建一个初始值为1的target
	target := big.NewInt(1)
	//左移256-bits位
	target = target.Lsh(target, 256-conf.TargetBit)
	return &ProofOfWork{block, target}
}

//根据block生成一个byte数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.Data,
			utils.IntToHex(pow.Block.Height),
			utils.IntToHex(int64(pow.Block.TimeStamp)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

//挖矿并且返回有效的哈希和nonce值
func (pow *ProofOfWork) Run() ([]byte, int64) {
	//将Block的属性拼接成字节数组
	//生成Hash
	//循环判断Hash的有效性，满足条件，跳出循环结束验证
	nonce := 0
	//用于存储新生成的hash
	hashInt := new(big.Int)
	var hash [32]byte
	for {
		//获取字节数组
		dataBytes := pow.prepareData(nonce)
		//生成hash
		hash = sha256.Sum256(dataBytes)
		// 不断的计算
		fmt.Printf("\r%d: %x", nonce, hash)
		//将hash存储到hashInt
		hashInt.SetBytes(hash[:])
		/*
			判断hashInt是否小于Block里的target
			Com compares x and y and returns:
			-1 if x < y
			0 if x == y
			1 if x > y
		*/
		if pow.Target.Cmp(hashInt) == 1 {
			break
		}
		nonce++
	}
	fmt.Println()
	return hash[:], int64(nonce)
}

// 判断算出来的hash值是否有效
func (pow *ProofOfWork) IsValid() bool {
	hashInt := new(big.Int)
	hashInt.SetBytes(pow.Block.Hash)
	return pow.Target.Cmp(hashInt) == 1
}
