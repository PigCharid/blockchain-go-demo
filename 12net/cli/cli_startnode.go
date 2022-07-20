package cli

import (
	"fmt"
	"os"
	"publicchain/server"
	"publicchain/wallet"
)

// 启动节点服务
func (cli *CLI) startNode(nodeID string, minerAdd string) {
	// 启动服务器
	fmt.Println(nodeID, minerAdd)
	if minerAdd == "" || wallet.IsValidForAddress([]byte(minerAdd)) {
		//  启动服务器
		fmt.Printf("启动服务器:localhost:%s\n", nodeID)
		server.StartServer(nodeID, minerAdd)

	} else {
		fmt.Println("指定的地址无效")
		os.Exit(0)
	}

}
