package pbcc

//输出结构体
type TXOuput struct {
	Value        int64  // 就是币的数量
	ScriptPubKey string //公钥：先理解为，用户名
}

//判断当前txOutput消费，和指定的address是否一致
func (txOutput *TXOuput) UnLockWithAddress(address string) bool {
	return txOutput.ScriptPubKey == address
}
