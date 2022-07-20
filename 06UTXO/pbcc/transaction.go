package pbcc

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"publicchain/utils"
	"time"
)

//Transaction结构体
type Transaction struct {
	TxID  []byte     //交易ID
	Vins  []*TXInput //输入
	Vouts []*TXOuput //输出
}

// 铸币交易
func NewCoinBaseTransaction(address string) *Transaction {
	txInput := &TXInput{[]byte{}, -1, "coinbase Data"}
	txOutput := &TXOuput{10, address}
	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOuput{txOutput}}
	//设置hash值
	txCoinbase.SetTxID()
	return txCoinbase
}

//设置交易的hash
func (tx *Transaction) SetTxID() {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	buffBytes := bytes.Join([][]byte{utils.IntToHex(time.Now().Unix()), buff.Bytes()}, []byte{})
	hash := sha256.Sum256(buffBytes)
	tx.TxID = hash[:]
}

//判断当前交易是否是Coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxID) == 0 && tx.Vins[0].Vout == -1
}
