package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
)

//CLI结构体
type CLI struct {
}

//添加Run方法
func (cli *CLI) Run() {
	//判断命令行参数的长度
	isValidArgs()

	//创建flagset标签对象
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)

	//设置标签后的参数

	flagCreateBlockChainData := createBlockChainCmd.String("data", "Genesis block data", "创世区块交易数据")

	//解析
	switch os.Args[1] {

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

	default:
		fmt.Println("请检查参数的输入")
		printUsage()
		os.Exit(1)
	}

	if printChainCmd.Parsed() {
		cli.printChains()
	}

	if createBlockChainCmd.Parsed() {
		if *flagCreateBlockChainData == "" {
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockChainData)
	}

}

// 检查是否有命令行的参数
func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

// 打印提示
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateblockchain -data DATA -- 创建创世区块")
	fmt.Println("\tprintchain -- 输出信息")
}
