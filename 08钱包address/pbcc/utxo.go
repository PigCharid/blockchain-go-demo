package pbcc

//结构体UTXO，用于表示未花费的钱
type UTXO struct {
	TxID   []byte   //当前Transaction的交易ID
	Index  int      //这个在交易的输出的下标索引
	Output *TXOuput //输出
}
