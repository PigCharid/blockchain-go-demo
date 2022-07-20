package pbcc

import (
	"bytes"
	"publicchain/crypto"
)

//输出结构体
type TXOuput struct {
	Value      int64  //面值
	PubKeyHash []byte // 公钥
}

//判断当前txOutput消费，和指定的address是否一致
func (txOutput *TXOuput) UnLockWithAddress(address string) bool {
	fullPayloadHash := crypto.Base58Decode([]byte(address))
	pubKeyHash := fullPayloadHash[1 : len(fullPayloadHash)-4]
	return bytes.Equal(txOutput.PubKeyHash, pubKeyHash)
}

// 创建新的输出
func NewTXOuput(value int64, address string) *TXOuput {
	txOutput := &TXOuput{value, nil}
	txOutput.Lock(address)
	return txOutput
}

// 把输出锁定给某地址
func (txOutput *TXOuput) Lock(address string) {
	publicKeyHash := crypto.Base58Decode([]byte(address))
	txOutput.PubKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]
}
