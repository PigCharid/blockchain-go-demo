package cli

import (
	"fmt"
	"publicchain/wallet"
)

// 打印所有钱包地址
func (cli *CLI) addressLists(nodeID string) {
	fmt.Println("打印所有的钱包地址")
	//获取
	Wallets := wallet.NewWallets(nodeID)
	for address, _ := range Wallets.WalletsMap {
		fmt.Println("address:", address)
	}
}
