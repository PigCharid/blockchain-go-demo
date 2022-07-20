package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"publicchain/conf"
	"publicchain/pbcc"
	"publicchain/utils"
)

// 判断节点是否为已知节点
func nodeIsKnown(addr string) bool {
	for _, node := range KnowNodes {
		if node == addr {
			return true
		}
	}
	return false
}

// 启动一个节点服务
func StartServer(nodeID string, minerAdd string) {
	// 当前节点的IP地址
	NodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	// 旷工地址
	MinerAddress = minerAdd
	fmt.Printf("nodeAddress:%s,minerAddress:%s\n", NodeAddress, MinerAddress)
	// 和主节点建立起链接
	ln, err := net.Listen(conf.PROTOCOL, NodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	bc := pbcc.GetBlockchainObject(nodeID)
	// 第一个终端：端口为8000,启动的就是主节点
	// 第二个终端：端口为8001，钱包节点
	// 第三个终端：端口号为8002，矿工节点
	if NodeAddress != KnowNodes[0] {
		// 此节点是钱包节点或者矿工节点，需要向主节点发送请求同步数据
		fmt.Printf("主节点是:%s\n", KnowNodes[0])
		SendVersion(KnowNodes[0], bc)
	}
	for {
		// 收到的数据的格式是固定的，12字节+结构体字节数组
		// 接收客户端发送过来的数据
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		// go出去处理发来的消息
		go handleConnection(conn, bc)
	}
}

// 处理节点连接中的数据请求
func handleConnection(conn net.Conn, bc *pbcc.BlockChain) {
	// 读取客户端发送过来的所有的数据
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("收到的消息类型是:%s\n", request[:conf.COMMANDLENGTH])
	//获取消息类型
	command := utils.BytesToCommand(request[:conf.COMMANDLENGTH])
	switch command {
	case conf.COMMAND_VERSION:
		handleVersion(request, bc)

	case conf.COMMAND_GETBLOCKS:
		handleGetblocks(request, bc)

	case conf.COMMAND_INV:
		handleInv(request, bc)

	case conf.COMMAND_ADDR:
		//handleAddr(request, bc)  预留一个地址处理可以当作业自己去发挥一下
	case conf.COMMAND_BLOCK:
		handleBlock(request, bc)

	case conf.COMMAND_GETDATA:
		handleGetData(request, bc)

	case conf.COMMAND_TX:
		handleTx(request, bc)
	default:
		fmt.Println("未知消息类型")
	}
	conn.Close()
}
