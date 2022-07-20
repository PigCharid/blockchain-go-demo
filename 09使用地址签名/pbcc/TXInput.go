package pbcc

import (
	"bytes"
	"publicchain/wallet"
)

//输入结构体
type TXInput struct {
	TxID      []byte //交易的ID
	Vout      int    //存储Txoutput的vout里面的索引
	Signature []byte //数字签名
	PublicKey []byte //公钥
}

//判断当前txInput消费，和指定的address是否一致
func (txInput *TXInput) UnLockWithAddress(pubKeyHash []byte) bool {
	//把交易里面的公钥进行计算
	publicKey := wallet.PubKeyHash(txInput.PublicKey)
	return bytes.Equal(pubKeyHash, publicKey)
}
