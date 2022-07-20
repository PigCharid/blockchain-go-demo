package pbcc

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TxOutputs struct {
	UTXOS []*UTXO
}

//序列化
func (outs *TxOutputs) Serilalize() []byte {
	//创建一个buffer
	var result bytes.Buffer
	//创建一个编码器
	encoder := gob.NewEncoder(&result)
	//编码--->打包
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TxOutputs {
	var txOutputs TxOutputs
	var reader = bytes.NewReader(txOutputsBytes)
	//创建一个解码器
	decoder := gob.NewDecoder(reader)
	//解包
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}
	return &txOutputs
}
