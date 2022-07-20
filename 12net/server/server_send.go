package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"publicchain/conf"
	"publicchain/pbcc"
	"publicchain/utils"
)

// 像其他节点发送数据
func SendData(to string, data []byte) {
	// 获取链接对象
	conn, err := net.Dial(conf.PROTOCOL, to)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 附带要发送的数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

//组装版本消息数据并发送
func SendVersion(toAddress string, blc *pbcc.BlockChain) {
	// 获取获取区块高度
	bestHeight := blc.GetBestHeight()
	// 组装版本数据
	payload := utils.GobEncode(Version{conf.NODE_VERSION, bestHeight, NodeAddress})
	// 把命令和数据组成请求
	request := append(utils.CommandToBytes(conf.COMMAND_VERSION), payload...)
	fmt.Printf("节点%s向节点%s发送了version消息\n", NodeAddress, toAddress)
	// 数据发送
	SendData(toAddress, request)
}

//组装获取区块消息并发送
func SendGetBlocks(toAddress string) {
	// 指定从全节点获取
	payload := utils.GobEncode(GetBlocks{NodeAddress})
	// 拼接命令和数据
	request := append(utils.CommandToBytes(conf.COMMAND_GETBLOCKS), payload...)
	fmt.Printf("向节点地址为:%s的节点发送了GetBlock消息\n", toAddress)
	SendData(toAddress, request)
}

// 组装Inv消息并发送
func SendInv(toAddress string, kind string, hashes [][]byte) {
	// 从全节点获取
	payload := utils.GobEncode(Inv{NodeAddress, kind, hashes})
	// 拼接命令和数据
	request := append(utils.CommandToBytes(conf.COMMAND_INV), payload...)
	fmt.Printf("节点%s向节点%s发送了Inv消息\n", NodeAddress, toAddress)
	SendData(toAddress, request)
}

// 组装GetData消息并发送
func SendGetData(toAddress string, kind string, blockHash []byte) {
	// 向全节点获取
	payload := utils.GobEncode(GetData{NodeAddress, kind, blockHash})
	request := append(utils.CommandToBytes(conf.COMMAND_GETDATA), payload...)
	fmt.Printf("节点%s向节点%s发送了GetData消息\n", NodeAddress, toAddress)
	SendData(toAddress, request)
}

// 组装BlockData消息并发送
func SendBlock(toAddress string, block []byte) {
	payload := utils.GobEncode(BlockData{NodeAddress, block})
	request := append(utils.CommandToBytes(conf.COMMAND_BLOCK), payload...)
	fmt.Printf("节点%s向节点%s发送了Block消息\n", NodeAddress, toAddress)
	SendData(toAddress, request)
}

// 组装TXData消息并发送
func SendTx(toAddress string, tx *pbcc.Transaction) {
	payload := utils.GobEncode(Tx{NodeAddress, tx})
	request := append(utils.CommandToBytes(conf.COMMAND_TX), payload...)
	fmt.Printf("节点%s向节点%s发送了Tx消息\n", NodeAddress, toAddress)
	SendData(toAddress, request)
}
