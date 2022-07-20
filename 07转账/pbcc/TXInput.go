package pbcc

//输入结构体
type TXInput struct {
	TxID      []byte //交易的ID
	Vout      int    //存储Txoutput的vout里面的索引
	ScriptSiq string //用户名
}

//判断当前txInput消费，和指定的address是否一致
func (txInput *TXInput) UnLockWithAddress(address string) bool {
	return txInput.ScriptSiq == address
}
