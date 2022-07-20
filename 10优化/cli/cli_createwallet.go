package cli

import "publicchain/wallet"

// 创建一个新钱包地址
func (cli *CLI) createWallet() {
	wallets := wallet.NewWallets()
	wallets.CreateNewWallet()
}
