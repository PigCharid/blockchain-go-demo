package pbcc

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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

// 创建普通交易
func NewSimpleTransaction(from, to string, amount int64, bc *BlockChain, txs []*Transaction) *Transaction {
	var txInputs []*TXInput
	var txOutputs []*TXOuput
	// 未打包的中找到够花的utxo
	balance, spendableUTXO := bc.FindSpendableUTXOs(from, amount, txs)
	// 遍历spendableUTXO来组装
	for txID, indexArray := range spendableUTXO {
		txIDBytes, _ := hex.DecodeString(txID)
		for _, index := range indexArray {
			txInput := &TXInput{txIDBytes, index, from}
			txInputs = append(txInputs, txInput)
		}
	}
	//转账
	txOutput1 := &TXOuput{amount, to}
	txOutputs = append(txOutputs, txOutput1)
	// 找零
	txOutput2 := &TXOuput{balance - amount, from}
	txOutputs = append(txOutputs, txOutput2)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.SetTxID()
	return tx
}
