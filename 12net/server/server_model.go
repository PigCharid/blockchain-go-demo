package server

import "publicchain/pbcc"

//version消息结构体
type Version struct {
	Version    int64  // 版本
	BestHeight int64  // 当前节点区块的高度
	AddrFrom   string //当前节点的地址
}

//请求区块信息结构 意为 “给我看一下你有什么区块”（在比特币中，这会更加复杂）
type GetBlocks struct {
	AddrFrom string //对方的节点地址
}

// Inv消息结构体 像别人展示自己的区块或者交易的信息
type Inv struct {
	AddrFrom string   //自己的地址
	Type     string   //类型 block tx
	Items    [][]byte //hash二维数组
}

// GetData消息结构体  用于某个块或交易的请求，它可以仅包含一个块或交易的ID。
type GetData struct {
	AddrFrom string
	Type     string
	Hash     []byte //获取的是hash
}

// BlockData消息结构体 给发送GetData请求回复区块
type BlockData struct {
	AddrFrom string
	Block    []byte
}

// Tx消息结构体 给发送GetData请求回复交易
type Tx struct {
	AddrFrom string
	Tx       *pbcc.Transaction
}
