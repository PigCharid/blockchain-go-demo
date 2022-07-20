package cli

import "publicchain/wallet"

// 创建一个新钱包地址
func (cli *CLI) createWallet(nodeID string) {
	wallets := wallet.NewWallets(nodeID)
	wallets.CreateNewWallet(nodeID)
}
