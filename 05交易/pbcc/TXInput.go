package pbcc

//输入结构体
type TXInput struct {
	TxID      []byte //交易的ID
	Vout      int    //索引  之前交易输出的下标索引
	ScriptSiq string //用户名
}
