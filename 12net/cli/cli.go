package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"publicchain/utils"
	"publicchain/wallet"
)

//CLI结构体
type CLI struct {
}

//Run方法
func (cli *CLI) Run() {
	//判断命令行参数的长度
	isValidArgs()
	// 通过环境变量获取
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID环境变量没有设置\n")
		os.Exit(1)
	}
	fmt.Printf("本节点的NODE_ID是:%s\n", nodeID)

	//创建flagset标签对象
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	addressListsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	//设置标签后的参数
	flagFromData := sendBlockCmd.String("from", "", "转帐源地址")
	flagToData := sendBlockCmd.String("to", "", "转帐目标地址")
	flagAmountData := sendBlockCmd.String("amount", "", "转帐金额")
	flagCreateBlockChainData := createBlockChainCmd.String("address", "", "创世区块交易地址")
	flagGetBalanceData := getBalanceCmd.String("address", "", "要查询的某个账户的余额")
	flagMiner := startNodeCmd.String("miner", "", "定义挖矿奖励的地址")
	flagMine := sendBlockCmd.Bool("mine", false, "是否在当前节点中立即验证")

	//解析
	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslists":
		err := addressListsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1) //退出
	}

	if sendBlockCmd.Parsed() {
		if *flagFromData == "" || *flagToData == "" || *flagAmountData == "" {
			printUsage()
			os.Exit(1)
		}
		from := utils.JSONToArray(*flagFromData)
		to := utils.JSONToArray(*flagToData)
		amount := utils.JSONToArray(*flagAmountData)

		for i := 0; i < len(from); i++ {
			if !wallet.IsValidForAddress([]byte(from[i])) || !wallet.IsValidForAddress([]byte(to[i])) {
				fmt.Println("钱包地址无效")
				printUsage()
				os.Exit(1)
			}
		}
		cli.send(from, to, amount, nodeID, *flagMine)
	}
	if printChainCmd.Parsed() {
		cli.printChains(nodeID)
	}
	if createBlockChainCmd.Parsed() {
		if !wallet.IsValidForAddress([]byte(*flagCreateBlockChainData)) {
			fmt.Println("创建地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockChainData, nodeID)
	}

	if getBalanceCmd.Parsed() {
		//if *flagGetBalanceData == "" {
		if !wallet.IsValidForAddress([]byte(*flagGetBalanceData)) {
			fmt.Println("查询地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceData, nodeID)
	}

	if createWalletCmd.Parsed() {
		//创建钱包
		cli.createWallet(nodeID)
	}

	//获取所有的钱包地址
	if addressListsCmd.Parsed() {
		cli.addressLists(nodeID)
	}

	if testCmd.Parsed() {
		cli.TestMethod(nodeID)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(nodeID, *flagMiner)
	}

}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreatewallet -- 创建钱包")
	fmt.Println("\taddresslists -- 输出所有钱包地址")
	fmt.Println("\tcreateblockchain -address DATA -- 创建创世区块")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -mine -- 交易明细.")
	fmt.Println("\tprintchain - 输出信息:")
	fmt.Println("\tgetbalance -address DATA -- 查询账户余额")
	fmt.Println("\ttest -- 测试")
	fmt.Println("\tstartnode -miner ADDRESS -- 启动节点服务器，并且指定挖矿奖励的地址.")
}
