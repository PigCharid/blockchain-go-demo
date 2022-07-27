# 前言

之前在github看到一个用go实现的简单的区块链系统，现在我自己也来整理一个，该教程中，很更加详细的去验证和解释很多参数和概念，也添加了大量的注释，同时对很多场景也都进行了测试，喜欢的朋友可以给个star权当鼓励，也可以一起探讨区块链相关知识，我们共同进步

本教程会从区块链的基础实现开始，然后逐步的深入到以太坊源码的分析，后面也会对web3方向的知识出一些教程

本教程的代码是不断的在原有的代码基础上不断的更新迭代，每一个章节的代码也单独列了出来，很多知识在本教程中没有详细的展开，但是代码标注的很详细，大家可以在代码层面加以理解，这样更能够深入理解

作者微信：`13721072141`

# 一、基础结构

## 1、搭建项目

在gopath下创建项目publicchain

创建项目目录，下面是项目最终的包结构

```
├── cli				//命令行参数
├── conf			//配置参数
├── crypto		//加密工具
├── main.go		//主函数
├── pbcc			//publicchain核心部分
├── server		//p2p网络服务
├── utils			//工具包
└── wallet		//钱包
```

创建完包的架构目录后

继续在终端运行`go mod init`，此时文件夹中将会多出一个go.mod文

## 2、区块

区块链以区块（block）的形式储存数据信息，一个区块记录了一段时间内系统或网络中产生的重要数据信息，区块通过引用上一个区块的hash值来连接上一个区块这样区块就按时间顺序排列形成了一条链。每个区块应该包含头部（head）信息用于总结性的描述这个区块，然后在区块的数据存放区（body）中存放要保存的重要数据，在我们实现的区块中，就没有对区块进行头部和数据存放区的划分

关于区块的其他知识，这边就不在过多的赘述

在pbcc包下创建block.go，我们先来构建区块结构体

`block.go`

```go
//Block结构体
type Block struct {
	Height        int64  //高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	PrevBlockHash []byte //上一个区块的哈希值ProvHash：
	Data          []byte //交易数据Data：目前先设计为[]byte,后期是Transaction
	TimeStamp     int64  //时间戳TimeStamp：
	Hash          []byte //哈希值Hash：32个的字节，64个16进制数
}
```

我们定义的区块中有区块高度，上一个区块的hash，交易数据，时间戳，本身的哈希值

实际上时间戳，本身的哈希值，指向上一个区块的哈希这三个属性构成头部信息，而区块中的数据以Data属性表示，其实在比特币系统里面数据是以交易的形式来充当的，我们暂时使用一个字节数组来当数据

每个区块都要自己的hash值，那么我们接下来就来设置区块的hash值

首先我们需要一些辅助工具，接下来在项目下创建一个uitls包，在包下创建utils.go

因为区块的高度是用int64类型的，提供一个int64转化为byte数组的辅助工具

`utils.go`

```go
//将int64转换为bytes
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
```

这里区块高度采用的是int64，也就是64位，那么我们要把它转化成`[]byte`字节数组，8个字节就够用了，那么转化成的结果是一个8位的字节数组，我们在main.go测试一下看一下

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/utils"
)
func main() {
	intbyte := utils.IntToHex(256)
	fmt.Println(intbyte)
}

```

```
输出结果
[0 0 0 0 0 0 1 0]
```

看到结果以后我们就很好理解int64转化成[]byte字节数组是什么样的了



那么接下来就可以设置当前区块的hash了

`block.go`

```go
//设置区块的hash
func (block *Block) SetHash() {
	//将高度转为字节数组
	heightBytes := utils.IntToHex(block.Height)
	// 先将时间戳按二进制转化成字符串
	timeString := strconv.FormatInt(block.TimeStamp, 2)
	// 强转一下
	timeBytes := []byte(timeString)
	//拼接所有的属性 把几个属性按照下面的空字节来分割拼接
	blockBytes := bytes.Join([][]byte{
		heightBytes,
		block.PrevBlockHash,
		block.Data,
		timeBytes}, []byte{})
	//生成哈希值，返回一个32位的字节数组 256位
	//fmt.Println(blockBytes)
	hash := sha256.Sum256(blockBytes)
	block.Hash = hash[:]
}
```

在设置区块的hash的接口中，主要分成了三步骤

​		1.将高度，时间戳转换为字节数组

​		2.拼接所有属性

​		3.将拼接后的字节数组转换为Hash值

接下来我们一起来看下拼接后得到的结果

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/pbcc"
	"time"
)

func main() {
	block := pbcc.Block{
		Height:        1,
		PrevBlockHash: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		Data:          []byte{0, 0, 0, 0, 0, 0, 0, 2},
		TimeStamp:     time.Now().Unix(),
	}
	block.SetHash()
	fmt.Println(block)
}
```

我们手动创建了一个区块高度为1，上一个区块hash为0，数据为00000002，时间戳为当前时间的区块，然后对这个区块使用我们刚写的设置区块的hash的接口方法

同时在SetHash()的时候输出一下拼接完各个字段的结果

```
输出结果
[0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 49 49 48 48 48 49 48 49 49 48 48 49 48 49 48 49 49 49 49 49 48 49 49 49 48 48 49 48 49 48 48]
{1 [0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 2] 1657469844 [226 103 181 149 108 205 190 10 54 20 43 145 151 174 20 63 3 91 28 195 202 110 63 173 231 75 0 129 242 107 245 236]}
```

可以看到区块的前四个字段拼接得到的结果和预期的一样，区块最后一个字段的信息是区块的hash

我们已经可以设置一个区块的全部信息了，接下来提供一个生成新区块的接口方法

`block.go`

```go
func NewBlock(data string, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, []byte(data), time.Now().Unix(), nil}
	//设置哈希值
	block.SetHash()
	return block
}
```

目前使我们是通过传入区块的数据，上一个区块的hash，区块高度来构建一个区块，这是目前用的方法，当然我们现在创建的是最简单的区块结构，后面我们会对区块的结构不断的升级

区块结构体和创建新区块的接口方法也有了，那么在区块链中有个很重要的知识，那么就是创世区块，接下来提供一个生成创世区块的方法

`block.go`

```go
//创建创世区块：
func CreateGenesisBlock(data string) *Block {
	return NewBlock(data, make([]byte, 32), 0)
}
```

我们指定创世区块的区块高度为0，上一个区块的hash为一个空的32位字节数组



基本的区块结构我们已经设置完成，下面创建main.go测试一下

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/pbcc"
)

func main() {
	genesisblock := pbcc.CreateGenesisBlock("i am genesisblock")
	fmt.Println(genesisblock)
}

```

```
输出结果
&{0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] [105 32 97 109 32 103 101 110 101 115 105 115 98 108 111 99 107] 1657470432 [49 62 253 196 172 112 160 20 155 148 171 139 74 211 178 16 81 16 226 192 9 149 190 134 81 69 110 135 230 96 211 211]}
```

输出了创世区块的区块高度，上一个区块的hash，数据，时间戳和当前区块的hash，其中上一个区块的hash、数据和当前区块的hash都是以字节数组的形式，通过结果中上一个区块的hash的字段，可以发现该区块就是一个创世区块

## 3、区块链

在获得了区块后，我们可以定义区块链，区块链就是区块的一个集合，区块“链“链起来的原理就是靠当前区块记录上一个区块的hash，这样就能够把所有的区块联系到一起了

下面我们在pbcc包下创建blockchain.go

我们目前先让所有区块放进一个切片里面构成区块链，这个只是暂时的构建区块链的方法

`blockchain.go`

```go
//创建区块链
type BlockChain struct {
	Blocks []*Block
}
```

每一个区块链中，第一个区块就是创世区块，也是也就是说每一个区块链都会含有创世快，下面提供一个创建区块链并带有创世区块的方法接口

`blockchain.go`

```go
//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(data string) *BlockChain {
	//创建创世区块
	genesisBlock := CreateGenesisBlock(data)
	//返回区块链对象
	return &BlockChain{[]*Block{genesisBlock}}
}
```

现在已经可以得到带有创世区块的链了，那么接下来就是要往链里面添加新的区块

`blockchain.go`

```go
//添加一个新的区块，到区块链中
func (bc *BlockChain) AddBlockToBlockChain(data string, height int64, prevHash []byte) {
	//创建新区块
	newBlock := NewBlock(data, prevHash, height)
	//添加到切片中
	bc.Blocks = append(bc.Blocks, newBlock)
}
```

下面来测试一下是否能够得到一个区块链，并且往里面添加新的区块

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/pbcc"
)

func main() {
	//创建带有创世区块的区块链
	blockchain := pbcc.CreateBlockChainWithGenesisBlock("i am genesisblock")
	//添加一个新区快
	blockchain.AddBlockToBlockChain("first Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("second Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("third Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	for _, block := range blockchain.Blocks {
		fmt.Printf("Timestamp: %d\n", block.TimeStamp)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("height: %d\n", block.Height)
		fmt.Println("--------------------------------------------")
	}

```

我们看到运行的结果，打印的内容为包含创世区块在内的四个区块的区块链，注意观察，这些区块的hash和区块中上一个区块hash字段，这时候我们发现，最简单的区块链我们已经完成了

```
运行结果
Timestamp: 1657470808
hash: ad5c82c899b6d93030cc308f1112dfee9dd1ffd945c4e55366d744ba935095fc
Previous hash: 0000000000000000000000000000000000000000000000000000000000000000
data: i am genesisblock
height: 0
--------------------------------------------
Timestamp: 1657470808
hash: f5eb5c4c9d67ed809c6ea1364c0bc4fd6e7a6a51ea568377faa4305bdd526809
Previous hash: ad5c82c899b6d93030cc308f1112dfee9dd1ffd945c4e55366d744ba935095fc
data: first Block
height: 1
--------------------------------------------
Timestamp: 1657470808
hash: 5ce480b74737af1c228e2469c2ac16cc3143fb4c9702e4c7f8f2d570de2f1b8c
Previous hash: f5eb5c4c9d67ed809c6ea1364c0bc4fd6e7a6a51ea568377faa4305bdd526809
data: second Block
height: 2
--------------------------------------------
Timestamp: 1657470808
hash: 931aa3d9c33bef2d60fb2704fe1d0ac3809ce233bfcf0f7bb6afc1e27fc4583c
Previous hash: 5ce480b74737af1c228e2469c2ac16cc3143fb4c9702e4c7f8f2d570de2f1b8c
data: third Block
height: 3
--------------------------------------------
```

# 二、共识

我们常说区块链是一个分布式系统，系统中每个节点都有机会储存数据信息构造一个区块然后追加到区块链尾部。这里就存在一个问题，那就是当区块链系统中有多个节点都想将自己的区块追加到区块链是我们该怎么办？我们将这些等待添加的区块统称为候选区块，显然我们不能对候选区块全盘照收，否则区块链就不再是一条链而是不同分叉成区块树。那么我们如何确定一种方法来从候选区块中选择一个加入到区块链中了？这里就需要用到区块链的共识机制，后文将以比特币使用的最经典PoW共识机制进行讲解

共识机制说的通俗明白一点就是要在相对公平的条件下让想要添加区块进区块链的节点内卷，通过竞争选择出一个大家公认的节点添加它的区块进入区块链。整个共识机制被分为两部分，首先是竞争，然后是共识。中本聪在比特币中设计了如下的一个Game来实现竞争：每个节点去寻找一个随机值（也就是nonce），将这个随机值作为候选区块的头部信息属性之一，要求候选区块对自身信息（注意这里是包含了nonce的）进行哈希后表示为数值要小于一个难度目标值（也就是Target），最先寻找到nonce的节点即为卷王，可以将自己的候选区块发布并添加到区块链尾部。这个Game设计的非常巧妙，首先每个节点要寻找到的nonce只对自己候选区块有效，防止了其它节点同学抄答案；其次，nonce的寻找是完全随机的没有技巧，寻找到nonce的时间与目标难度值与节点本身计算性能有关，但不妨碍性能较差的节点也有机会获胜；最后寻找nonce可能耗费大量时间与资源，但是验证卷王是否真的找到了nonce却非常却能够很快完成并几乎不需要耗费资源，这个寻找到的nonce可以说就是卷王真的是卷王的证据。现在我们就来一步一步实现这个Game

## 1、区块新字段Nonce

我们都知道一个合法区块的诞生其哈希值必须满足指定的条件，比特币采用的是工作量证明。我们这里用go开发的公链也采用POW一致性算法来产生合法性区块

因此，区块必须不断产生哈希直到满足POW的哈希值产生才能添加到主链上成为合法区块，对于一个区块来说，1-4项属性都是固定的，（高度、上一个区块的hash、数据、时间都是不能改变的）而区块哈希又是由这些属性拼接取哈希生成的。所以，要想让区块哈希能不断变化，必须引入一个变量Nonce

引入Nonce后，就可以通过改变Nonce值来不断产生新的哈希值直到找到满足条件的哈希

修改block.go下的block结构体

`block.go`

```go
//Block结构体
type Block struct {
	Height        int64  //高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	PrevBlockHash []byte //上一个区块的哈希值ProvHash：
	Data          []byte //交易数据Data：目前先设计为[]byte,后期是Transaction
	TimeStamp     int64  //时间戳TimeStamp：
	Hash          []byte //哈希值Hash：32个的字节，64个16进制数
	Nonce         int64  // 随机数
}
```

改完以后项目会出现报错，因为区块的结构新增了字段，但是现在我们不用管这个，我么先设置POW部分

## 2、POW工作量证明机制

对于256位的哈希值来说设定挖矿条件的方式往往是：前多少位为0，targetBits便是用于指定目标哈希需满足的条件的，即计算的哈希值必须前targetBits位为0，这个也是区块链中用于挖矿难度设定的一个参数

在配置文件中设置POW目标值的偏移

在conf包下创建config.go

`config.go`

```go
//256位Hash里面前面至少有16个零
const TargetBit = 16
```

接下来在pbcc包下创建proofOfwork.go

创建POW结构体

`proofOfwork.go`

```go
//pow结构体
type ProofOfWork struct {
	Block  *Block   //要验证的区块
	Target *big.Int //大整数存储,目标哈希
}
```

该结构两个字段，一个是区块，一个是目标值，那么就是不断的改变区块内数据的nonce值，然后计算哈希和目标值比较

POW的原理我们已经明白了，接下来提供一个创建新的工作量证明对象函数，把这个区块传入，我们这里设置的目标值就简单的把大整数1左移规定的位数

```go
//创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {
	/**
	target计算方式  假设：Hash为8位，targetBit为2位
	eg:0000 0001(8位的Hash)
	1.8-2 = 6 将上值左移6位
	2.0000 0001 << 6 = 0100 0000 = target
	3.只要计算的Hash满足 ：hash < target，便是符合POW的哈希值
	*/

	//创建一个初始值为1的target
	target := big.NewInt(1)
	//左移256-bits位
	target = target.Lsh(target, 256-conf.TargetBit)
	return &ProofOfWork{block, target}
}
```

有了工作量证明对象的话，就是区块和目标值都有了，那么就是不断的计算了

接下来第一步就是将对象里面区块信息连同nonce值一起打包，然后计算

先提供一个将区块字段进行封装成字节数组的方法

```go
//根据block生成一个byte数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.Data,
			utils.IntToHex(pow.Block.Height),
			utils.IntToHex(int64(pow.Block.TimeStamp)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}
```

拼接完区块的数据以后那么就可以去计算有效的hash值，这一个过程在区块链的术语就是`挖矿`

```go
func (pow *ProofOfWork) Run() ([]byte, int64) {
	//将Block的属性拼接成字节数组
	//生成Hash
	//循环判断Hash的有效性，满足条件，跳出循环结束验证
	nonce := 0
	//用于存储新生成的hash
	hashInt := new(big.Int)
	var hash [32]byte
	for {
		//获取字节数组
		dataBytes := pow.prepareData(nonce)
		//生成hash
		hash = sha256.Sum256(dataBytes)
		// 不断的计算
		fmt.Printf("\r%d: %x", nonce, hash)
		//将hash存储到hashInt
		hashInt.SetBytes(hash[:])
		/*
			判断hashInt是否小于Block里的target
			Com compares x and y and returns:
			-1 if x < y
			0 if x == y
			1 if x > y
		*/
		if pow.Target.Cmp(hashInt) == 1 {
			break
		}
		nonce++
	}
	fmt.Println()
	return hash[:], int64(nonce)
}
```

可以看到，神秘的nonce不过是从0开始取的整数而已，随着不断尝试，每次失败nonce就加1直到由当前nonce得到的区块哈希转化为数值小于目标难度值为止

当矿工计算出了有效的nonce广播给所有旷工以后，那么下面其他的旷工要做的一件事就是验证nonce值得有效性，那这里也是简单的把计算出来的hash和目标值进行比较

```go
// 判断算出来的hash值是否有效
func (pow *ProofOfWork) IsValid() bool {
	hashInt := new(big.Int)
	hashInt.SetBytes(pow.Block.Hash)
	return pow.Target.Cmp(hashInt) == 1
}
```

这时候已经完成了基础的proofOfwork，那么我们在创建新的区块的时候就需要让这个区块去做POW了

把创建区块的接口进行修改

```go
//创建新的区块
func NewBlock(data string, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, []byte(data), time.Now().Unix(), nil, 0}
	//调用工作量证明的方法，并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	// 然后把计算出来的结果赋给区块
	block.Hash = hash
	block.Nonce = nonce
	return block
}
```

可以看到我们在创建新的区块的时候，区块是经过了pow的计算，也就是说现在创建的新区块都是计算出了nonce值，在这个时候之前设置的设置区块的hash就不需要了

下面来测试一下proofOfwork

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/pbcc"
)

func main() {
	//创建带有创世区块的区块链
	blockchain := pbcc.CreateBlockChainWithGenesisBlock("i am genesisblock")
	//添加一个新区快
	blockchain.AddBlockToBlockChain("first Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("second Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockChain("third Block", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	for _, block := range blockchain.Blocks {
		fmt.Printf("Timestamp: %d\n", block.TimeStamp)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("height: %d\n", block.Height)
		fmt.Println("--------------------------------------------")
	}
}
```

运行后，consle会不断打印计算出的Hash值，直到计算出的Hash值满足条件

```go
结果
141665: 00005ebd52697468d4a0f64861b1848efb3597da2a7225617699a9d09f75289c
106640: 00009d1824f30b61df6bcfa4e8cf3e6cacdfa114154b47c47262b2a62255992b
266386: 0000bd68df1c99700ef8c9cf5dba4d270904582f227132811675a9ee37177458
129881: 000050335638289e2415878fb8e8bfd5ffd035272a56306ad3d924e3626815d9
Timestamp: 1657473133
hash: 00005ebd52697468d4a0f64861b1848efb3597da2a7225617699a9d09f75289c
Previous hash: 0000000000000000000000000000000000000000000000000000000000000000
data: i am genesisblock
height: 0
--------------------------------------------
Timestamp: 1657473137
hash: 00009d1824f30b61df6bcfa4e8cf3e6cacdfa114154b47c47262b2a62255992b
Previous hash: 00005ebd52697468d4a0f64861b1848efb3597da2a7225617699a9d09f75289c
data: first Block
height: 1
--------------------------------------------
Timestamp: 1657473141
hash: 0000bd68df1c99700ef8c9cf5dba4d270904582f227132811675a9ee37177458
Previous hash: 00009d1824f30b61df6bcfa4e8cf3e6cacdfa114154b47c47262b2a62255992b
data: second Block
height: 2
--------------------------------------------
Timestamp: 1657473150
hash: 000050335638289e2415878fb8e8bfd5ffd035272a56306ad3d924e3626815d9
Previous hash: 0000bd68df1c99700ef8c9cf5dba4d270904582f227132811675a9ee37177458
data: third Block
height: 3
--------------------------------------------
```

我们其实没有看到目标值和计算值，接下来我们就来探索一下POW的计算过程中的数据

我们在设置区块POW的目标值是下面的代码

```go
//创建一个初始值为1的target
	target := big.NewInt(1)
	//2.左移256-bits位
	target = target.Lsh(target, 256-conf.TargetBit)
```

然后我们设置的TargetBit是16

未移动之前target的值是00000000........1前面有255个0

左移240位 00000000000000010000000...前面15个0，后面240个0

移动完以后的字节变现形式就是[1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]

接下来就是去计算加入nonce以后的hash值，如果计算出来的值小于目标值，那么就计算成功

下面我们就设置一个区块，将nonce值设置为0，然后计算一次和目标值比较一下

`main.go`

```go
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"publicchain/conf"
	"publicchain/pbcc"
	"publicchain/utils"
	"time"
)

func main() {
	target := big.NewInt(1)
	//2.左移256-bits位
	target = target.Lsh(target, 256-conf.TargetBit)
	block := pbcc.Block{
		Height:        1,
		PrevBlockHash: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		Data:          []byte{0, 0, 0, 0, 0, 0, 0, 2},
		TimeStamp:     time.Now().Unix(),
		Nonce:         0,
	}

	data := bytes.Join(
		[][]byte{
			block.PrevBlockHash,
			block.Data,
			utils.IntToHex(block.Height),
			utils.IntToHex(int64(block.TimeStamp)),
			utils.IntToHex(int64(block.Nonce)),
		},
		[]byte{})
	hash := sha256.Sum256(data)
	fmt.Println(hash)
	hashInt := new(big.Int)
	hashInt.SetBytes(hash[:])
	if target.Cmp(hashInt) == 1 {
		fmt.Println("成功计算出了hash值")
	}
	fmt.Println("请修改nonce继续计算")
}
```

下面我们看一下执行的结果

```
[123 90 145 151 149 59 186 200 68 229 224 131 207 101 82 98 212 214 195 224 205 242 15 82 220 153 74 12 7 114 252 129]
请修改nonce继续计算
```

发现计算出来的hash值要大于目标值，需要重新的计算，可以用之前搭建不断修改nonce值最后算出哈希的方法来先算出nonce，然后把区块的数据和已经算出来的数据设置成一样，然后看下是否能够算出哈希，这个留给各位学习者自己尝试一下

# 三、链上数据持久化

区块链上也可以说是分布式的存储，那么数据肯定也需要持久化存储到节点中去

bitcoin客户端的区块信息是存储在LevelDB数据库中，我们既然要基于go开发公链，这里用到的数据库是基于go的boltDB

区块链的数据主要集中在各个区块上，所以区块链的数据持久化即可转化为对每一个区块的存储，boltDB是KV存储方式，因此这里我们可以以区块的哈希值为Key，区块为Value

此外，我们还需要存储最新区块的哈希值。这样，就可以找到最新的区块，然后按照区块存储的上个区块哈希值找到上个区块，以此类推便可以找到区块链上所有的区块

## 1、安装boltDB

在终端运行`go get github.com/boltdb/bolt`

关于DB的使用这里就不在做过多叙述

先在config.go添加DB的配置信息

```go
package conf

//256位Hash里面前面至少有16个零
const TargetBit = 16

const DBNAME = "blockchain.db"  //数据库名
const BLOCKTABLENAME = "blocks" //表名
```

## 2、区块数据的序列化和反序列化

我们知道，boltDB存储的键值对的数据类型都是字节数组，所以在存储区块前需要对区块进行序列化，当然读取区块的时候就需要做反序列化处理，那么接下来要提供一个对区块进行序列化和反序列化的方法

`block.go`

```go
//将区块序列化，得到一个字节数组---区块的行为
func (block *Block) Serilalize() []byte {
	//创建一个buffer
	var result bytes.Buffer
	//创建一个编码器
	encoder := gob.NewEncoder(&result)
	//编码--->打包
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
```

读取的时候就需要反序列化

```go
//反序列化，得到一个区块
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	var reader = bytes.NewReader(blockBytes)
	//创建一个解码器
	decoder := gob.NewDecoder(reader)
	//解包
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
```

## 3、区块链的重新定义

之前的区块链每次运行程序区块数组都是从零开始创建，并不能实现区块链的数据持久化。这里的数组属性要改为boltDB类型的区块数据库，同时还必须有一个存储当前区块链最新区块哈希的属性

每一个区块都会被存到数据库中，那么我们的区块链又改如何存储呢，其实只需要存放最新区块的hash和数据库对象就可以，只要有最新区块的hash，那么我们就可以凭借`PrevBlockHash`该字段一直往下寻找

那么区块链的结构我们进行修改

`blockchain.go`

```go
//创建区块链
type BlockChain struct {
	Tip []byte   // 最新区块的Hash值
	DB  *bolt.DB //数据库对象
}
```

这个时候如果导入boltDB包出错的话，那么运行`go mod tidy`

我们在创建区块链的时候就会直接去判断了，会先去判断一下当前有没有我们设置的KV数据库，接下来在utils.go先提供一个判断数据库是否存在的方法

`utils.go`

```go
//判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(conf.DBNAME); os.IsNotExist(err) {
		return false
	}
	return true
}
```

那么这个时候我们在创建区块链的时候就不是像之前一样每次都去生成一创世区块，而是先去检查本地的数据，如果没有的话那么再去创建带有创世区块的区块链，同时往数据库更新最新区块的hash

`blockchain.go`

```go
//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(data string) *BlockChain {
	//先判断数据库是否存在，如果有，从数据库读取
	if utils.DBExists() {
		fmt.Println("数据库已经存在")
		//打开数据库
		db, err := bolt.Open(conf.DBNAME, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		var blockchain *BlockChain
		//读取数据库
		err = db.View(func(tx *bolt.Tx) error {
			//打开表
			b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
			if b != nil {
				//读取最后一个hash
				hash := b.Get([]byte("l"))
				//创建blockchain
				blockchain = &BlockChain{hash, db}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		return blockchain
	}
	//数据库不存在，说明第一次创建，然后存入到数据库中
	fmt.Println("数据库不存在，创建中")
	//创建创世区块
	genesisBlock := CreateGenesisBlock(data)
	//打开数据库
	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(conf.BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误。。。")
			}
			//存储最新区块的hash
			b.Put([]byte("l"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//返回区块链对象
	return &BlockChain{genesisBlock.Hash, db}
}
```

之前新增区块的添加新区块的接口`func (blc *Blockchain) AddBlockToBlockchain(data string, height int64, prevHash []byte){}`

仔细看发现，参数好多显得巨繁琐。那是否有些参数是没必要传递的呢?

我们既然用数据库实现了区块链的数据持久化，这里的高度height可以根据上个区块高度自增，prevHash也可以从数据库中取出上个区块而得到。因此，从今天开始，该方法省去这两个参数，修改新增区块的方法

`blockchain.go`

```go
//添加一个新的区块，到区块链中
func (bc *BlockChain) AddBlockToBlockChain(data string) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//根据最新块的hash读取数据，并反序列化最后一个区块
			blockBytes := b.Get(bc.Tip)
			lastBlock := DeserializeBlock(blockBytes)
			//创建新的区块 根据最后一个区块注入prevblockhash和height
			newBlock := NewBlock(data, lastBlock.Hash, lastBlock.Height+1)
			//将新的区块序列化并存储
			err := b.Put(newBlock.Hash, newBlock.Serilalize())
			if err != nil {
				log.Panic(err)
			}
			//更新最后一个哈希值，以及blockchain的tip
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

我们不难发现区块链的区块遍历类似于单向链表的遍历，那么我们能不能制造一个像链表的Next属性似的迭代器，只要通过不断地访问Next就能遍历所有的区块

话都说到这份上了，答案当然是肯当的

在pbcc包下创建blockchainIterator.go

`blockchainIterator.go`

```go
//区块链迭代结构体
type BlockChainIterator struct {
	CurrentHash []byte   //当前区块的hash
	DB          *bolt.DB //数据库
}

//获取当前指向的区块，然后把指向改成上一个区块
func (bcIterator *BlockChainIterator) Next() *Block {
	block := new(Block)
	//打开数据库并读取
	err := bcIterator.DB.View(func(tx *bolt.Tx) error {
		//打开数据表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//根据当前hash获取数据并反序列化
			blockBytes := b.Get(bcIterator.CurrentHash)
			block = DeserializeBlock(blockBytes)
			//更新当前的hash
			bcIterator.CurrentHash = block.PrevBlockHash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block
}
```

接下来，我们有了迭代器，那么就可以利用迭代器来访问区块链

区块链对象中保存了最新区块的hash和数据库对象，迭代的话肯定是最新的区块开始，那么让区块链对象生成一个迭代器

`blockchain.go`

```go
//获取一个迭代器
func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.Tip, bc.DB}
}
```

下面利用迭代器打印区块链

```go
// 借助迭代器输出区块链
func (bc *BlockChain) PrintChains() {
	//获取迭代器对象
	bcIterator := bc.Iterator()
	//循环迭代
	for {
		block := bcIterator.Next()
		fmt.Printf("第%d个区块的信息:\n", block.Height+1)
		//获取当前hash对应的数据，并进行反序列化
		fmt.Printf("\t高度:%d\n", block.Height)
		fmt.Printf("\t上一个区块的hash:%x\n", block.PrevBlockHash)
		fmt.Printf("\t当前的hash:%x\n", block.Hash)
		fmt.Printf("\t数据:%s\n", block.Data)
		fmt.Printf("\t时间:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数:%d\n", block.Nonce)
		//直到父hash值为0
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}
```

数据持久化的设计也已经完成了，接下来测试一下

`main.go`

```go
package main

import "publicchain/pbcc"

func main() {
	blockchain := pbcc.CreateBlockChainWithGenesisBlock("i am genesis block")
	defer blockchain.DB.Close()

	//添加一个新区快
	blockchain.AddBlockToBlockChain("first Block")
	blockchain.AddBlockToBlockChain("second Block")
	blockchain.AddBlockToBlockChain("third Block")

	blockchain.PrintChains()
}

```

接下来创建区块链的时候先会创建创世区块

```
数据库不存在，创建中
121190: 00003862d2c467743b4d5ea9acbd30c58614c6040cb33d966144d05965243759
37863: 0000956acf28934ec19bc587273072cbf223c1ad3b6b047eb1c3c1e13b9e6c93
52343: 00002965afbcbfc2081fbc82732fea3ec17ed2ecf580cdde88b49a61743088bf
102941: 00008736d13f5cfa8f6fe5319ca94598222833e724edff55370e86b3c90328f7
第4个区块的信息:
        高度:3
        上一个区块的hash:00002965afbcbfc2081fbc82732fea3ec17ed2ecf580cdde88b49a61743088bf
        当前的hash:00008736d13f5cfa8f6fe5319ca94598222833e724edff55370e86b3c90328f7
        数据:third Block
        时间:2022-07-11 10:11:00
        次数:102941
第3个区块的信息:
        高度:2
        上一个区块的hash:0000956acf28934ec19bc587273072cbf223c1ad3b6b047eb1c3c1e13b9e6c93
        当前的hash:00002965afbcbfc2081fbc82732fea3ec17ed2ecf580cdde88b49a61743088bf
        数据:second Block
        时间:2022-07-11 10:10:58
        次数:52343
第2个区块的信息:
        高度:1
        上一个区块的hash:00003862d2c467743b4d5ea9acbd30c58614c6040cb33d966144d05965243759
        当前的hash:0000956acf28934ec19bc587273072cbf223c1ad3b6b047eb1c3c1e13b9e6c93
        数据:first Block
        时间:2022-07-11 10:10:57
        次数:37863
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:00003862d2c467743b4d5ea9acbd30c58614c6040cb33d966144d05965243759
        数据:i am genesis block
        时间:2022-07-11 10:10:53
        次数:121190
```

那么区块链的持久化存储目前就完成了

学习者们可以继续添加新的区块，并携带数据，然后打印，看下数据是否是在已经存在本地的区块链的上添加的还是新建的区块链

# 四、命令行工具

我们在使用bitcoin客户端或者以太坊客户端时，都是可以在命令行直接执行挖矿，新增交易等操作。这些操作都有赖于cli工具，我们写的基于go的公链也会支持这个功能

关于创建cli使用的的go下的flag包这里就不过多的叙述

创建cli包，在包下创建cli.go

```go
//CLI结构体
type CLI struct {
}

//添加Run方法
func (cli *CLI) Run() {
	//判断命令行参数的长度
	isValidArgs()
	//创建flagset标签对象
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	//设置标签后的参数
	flagAddBlockData := addBlockCmd.String("data", "helloworld", "交易数据")
	flagCreateBlockChainData := createBlockChainCmd.String("data", "Genesis block data", "创世区块交易数据")
	//解析
	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
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

	default:
		fmt.Println("请检查参数的输入")
		printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *flagAddBlockData == "" {
			printUsage()
			os.Exit(1)
		}
		cli.addBlock(*flagAddBlockData)
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
	fmt.Println("\taddblock -data Data -- 交易数据")
	fmt.Println("\tprintchain -- 输出信息")
}
```

目前可以进行创建区块链、添加新区块和输出区块链

在cli包下创建cli_addblock.go\cli_createblockchain.go\cli_print.go

对添加区块命令的处理

`cli_addblock.go`

```go
func (cli *CLI) addBlock(data string) {
	bc := pbcc.GetBlockchainObject()
	if bc == nil {
		fmt.Println("没有创世区块，无法添加")
		os.Exit(1)
	}
	defer bc.DB.Close()
	bc.AddBlockToBlockChain(data)
}
```

对创建区块链命令的处理

`cli_createblockchain.go`

```go
func (cli *CLI) createGenesisBlockchain(data string) {
	pbcc.CreateBlockChainWithGenesisBlock(data)
}
```

对打印区块链命令的处理

`cli_print.go`

```go
func (cli *CLI) printChains() {
	bc := pbcc.GetBlockchainObject()
	if bc == nil {
		fmt.Println("未创建，没有区块可以打印")
		os.Exit(1)
	}
	defer bc.DB.Close()
	bc.PrintChains()
}
```

上面对不同命令进行处理的时候，都会去获取最新的区块链数据，那么要提供一个获取最新区块链的方法

`blockchain.go`

```go
// 获取最新的区块链
func GetBlockchainObject() *BlockChain {
	if !utils.DBExists() {
		fmt.Println("数据库不存在，无法获取区块链")
		return nil
	}

	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var blockchain *BlockChain
	//读取数据库
	err = db.View(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//读取最后一个hash
			hash := b.Get([]byte("l"))
			//创建blockchain
			blockchain = &BlockChain{hash, db}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return blockchain
}
```

下面对cli进行测试

`main.go`

```go
 package main

import (
	"publicchain/cli"
)

func main() {
	//创建命令行工具
	cli := cli.CLI{}
	//激活cli
	cli.Run()
}

```

```
结果
go run main.go 
Usage:
        createblockchain -data DATA -- 创建创世区块
        addblock -data Data -- 交易数据
        printchain -- 输出信息
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createblockchain -data "i am genesis block"
数据库不存在，创建中
19319: 000013df4a9932fd9c08df0fab5077f8ed0ec4a472c8a1d3608db999bd2f285a
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go addblock -data "add block"
16236: 00009fc43190baabcc7849820a38a198411f8d26322a9fe072b4553711298149
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go printchain
第2个区块的信息:
        高度:1
        上一个区块的hash:000013df4a9932fd9c08df0fab5077f8ed0ec4a472c8a1d3608db999bd2f285a
        当前的hash:00009fc43190baabcc7849820a38a198411f8d26322a9fe072b4553711298149
        数据:add block
        时间:2022-07-11 10:33:59
        次数:16236
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:000013df4a9932fd9c08df0fab5077f8ed0ec4a472c8a1d3608db999bd2f285a
        数据:i am genesis block
        时间:2022-07-11 10:33:26
        次数:19319
```

这个时候我们已经有了一个简单的cli，当然，后面我们也会不断对这个进行完善

# 五、交易

区块链区块的作用是打包链上产生的交易,可以说交易是区块链至关重要的一个组成部分，在区块链中,交易一旦被创建，就没有任何人能够再去修改或是删除它，区块的数据部分就是由交易等信息组成

关于交易更多概念这里就不在做过多的叙述

## 1、交易

交易的结构其实是主要由交易的hash和输入输出构成，输入输出就是这个交易的来源，相当于来源和去处，一个交易的输入和输出可能会有多个，具体为什么会是这样的，后面介绍完UTXO的知识就明白了

在pbcc包下创建transaction.go

创建交易结构体

`transaction.go`

```go
//Transaction结构体
type Transaction struct {
	TxID  []byte     //交易ID
	Vins  []*TXInput //输入
	Vouts []*TXOuput //输出
}
```

在这里我们看到，交易的输入和输出都是以切片的形式出现的，那么说明一笔交易的输入和输出可能不是单一的

下面创建交易的输入和输出

在blockchain包下创建TXOutput.go和TXInput.go

首先我们要明白一件事，一笔交易输入的来源是之前已经发生的交易的输出，那么如果通过一个输入对象来锁定一笔钱呢，首先每一个交易都会有一个ID，然后通过对应输出的下标就能确定输出

来源确定好了，然后再添加一个花钱者的信息字段，就能构造出输入结构了

创建输入结构体

`TXInput.go`

```go
//输入结构体
type TXInput struct {
	TxID      []byte //交易的ID
	Vout      int    //存储Txoutput的vout里面的索引
	ScriptSiq string //用户名 
}
```

交易输入作为本次交易的消费源，输入来源于之前交易的输出，如上，TxHash是引用的上一笔输出所在的交易的交易哈希，Vout是该输出在相应交易中的输出索引，ScriptSig,，以暂时理解为用户名,表示哪一个用户拥有这一笔输入，ScriptSig的设定是为了保证用户只能话费自己名下的代币

创建输出结构体

`TXOutput.go`

```go
//输出结构体
type TXOuput struct {
	Value        int64  // 就是币的数量
	ScriptPubKey string //公钥：先理解为，用户名
}
```

这里的交易输出就是上面交易输入里引用的输出.Value是该输出的面值,ScriptPubKey暂时理解为用户名,表示谁将拥有这笔输出.



了解比特币的人都知道,交易输出是一个完整的不可分割的结构，什么意思呢？就是我们在引用输出作为输入时，,必须全部引用，不能仅仅使用其一部分，举个简单的🌰:

假如你有一个25btc的TXOutput，你需要花费10btc，这个交易的过程并不是你花费了25btc中的10btc，你的原有TXOutput依旧有15btc的余额。真正的过程是，你花费了整个原有的TXOutput，由于消费额不匹配,这里会产生一个15btc的找零，消费的结果是：你25btc的TXOutput被话费已不复存在，系统重新为你生成一个15btc面值的TXOutput，这两个TXOutput是完全不同的两个对象



## 2、区块添加交易字段

之前区块的数据是用一个`data []byte`来表示的，下面就对这个字段进行修改，数据字段我们将其设置成交易

`block.go`

```go
//Block结构体
type Block struct {
	Height        int64          //高度Height：其实就是区块的编号，第一个区块叫创世区块，高度为0
	PrevBlockHash []byte         //上一个区块的哈希值ProvHash：
	Txs           []*Transaction //交易数据Data：目前先设计为[]byte,后期是Transaction
	TimeStamp     int64          //时间戳TimeStamp：
	Hash          []byte         //哈希值Hash：32个的字节，64个16进制数
	Nonce         int64          // 随机数
}
```

由于区块的交易信息被修改了，那么一些接口要作出对应的修改

生成新的区块的时候，要把交易打包进区块，对创建区块进行修改

`block.go`

```go
//创建新的区块
func NewBlock(txs []*Transaction, provBlockHash []byte, height int64) *Block {
	//创建区块
	block := &Block{height, provBlockHash, txs, time.Now().Unix(), nil, 0}
	//调用工作量证明的方法，并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}
```

同时创建创世快也要修改

`block.go`

```go
//创建创世区块：
func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, make([]byte, 32), 0)
}
```

我们知道算区块的hash值或者说是POW在准备区块的字节数组的时候就需要把一个区块中传入的交易也全部转化成字节数组，我们只需要用所有交易的ID拼接字节数组就行，添加一个将所有交易信息转化成字节数组的接口，用于拼接字节数组

`block.go`

```go
//将Txs转为[]byte
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range block.Txs {
		txHashes = append(txHashes, tx.TxID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
```

其实还有这样一个问题，之前我们说一笔交易的输入是引入之前交易的输出作为来源，那么在区块链中，第一笔交易的输入来源于哪里呢？就像是先有鸡还是先有蛋的问题，那在区块链中其实还是要有一中特殊的交易，铸币交易，这种交易没有输入，只有输出，当旷工计算出新的区块的时候就会通过铸币交易给他们发送一笔钱

创建区块链的时候创建创世块中的交易要改为铸币交易

`blockchain.go`

```go
//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(address string) {
	if utils.DBExists() {
		fmt.Println("数据库已经存在")
		return
	}

	fmt.Println("数据库不存在,创建创世区块：")
	//先创建coinbase交易
	txCoinBase := NewCoinBaseTransaction(address)
	// 创世区块
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})
	//打开数据库
	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(conf.BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误")
			}
			//存储最新区块的hash
			b.Put([]byte("l"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

修改之前的添加区块带链上的接口，现在把新的区块添加到链上，需要提供交易数组

` blockchain.go`

```go
//添加一个新的区块，到区块链中
func (bc *BlockChain) AddBlockToBlockChain(txs []*Transaction) {
	//更新数据库
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//根据最新块的hash读取数据，并反序列化最后一个区块
			blockBytes := b.Get(bc.Tip)
			lastBlock := DeserializeBlock(blockBytes)
			//创建新的区块
			newBlock := NewBlock(txs, lastBlock.Hash, lastBlock.Height+1)
			//将新的区块序列化并存储
			err := b.Put(newBlock.Hash, newBlock.Serilalize())
			if err != nil {
				log.Panic(err)
			}
			//更新最后一个哈希值，以及blockchain的tip
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

修改打印区块链的信息，之前打印的交易是string，现在需要吧交易打印出来

`blockchain.go`

```go
// 借助迭代器输出区块链
func (bc *BlockChain) PrintChains() {
	//1.获取迭代器对象
	bcIterator := bc.Iterator()

	//2.循环迭代
	for {
		block := bcIterator.Next()
		fmt.Printf("第%d个区块的信息:\n", block.Height+1)
		//获取当前hash对应的数据，并进行反序列化
		fmt.Printf("\t高度:%d\n", block.Height)
		fmt.Printf("\t上一个区块的hash:%x\n", block.PrevBlockHash)
		fmt.Printf("\t当前的hash:%x\n", block.Hash)
		//fmt.Printf("\t数据：%v\n", block.Txs)
		fmt.Println("\t交易:")
		for _, tx := range block.Txs {
			fmt.Printf("\t\t交易ID:%x\n", tx.TxID)
			fmt.Println("\t\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\t\tTxID:%x\n", in.TxID)
				fmt.Printf("\t\t\tVout:%d\n", in.Vout)
				fmt.Printf("\t\t\tScriptSiq:%s\n", in.ScriptSiq)
			}
			fmt.Println("\t\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\t\tvalue:%d\n", out.Value)
				fmt.Printf("\t\t\tScriptPubKey:%s\n", out.ScriptPubKey)
			}
		}
		fmt.Printf("\t时间:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数/;%d\n", block.Nonce)
		//直到父hash值为0
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}
```

之前POW将区块信息打包成字节数组的时候是和data一起打包的，现在要修改成和交易的字节数组一起修改

```go
//根据block生成一个byte数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			utils.IntToHex(pow.Block.Height),
			utils.IntToHex(int64(pow.Block.TimeStamp)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}
```

删除命令行参数中的添加区块的命令设置，后面添加区块都是有交易发起的时候才会添加区块到链上了

## 3、CoinbaseTransaction

我们知道,当矿工成功挖到一个区块时会获得一笔奖励.那么这笔奖励是怎么交付到矿工账户的.这就有赖于一笔叫做铸币交易的交易，也很好理解，你是生成了一些奖励的币

铸币交易是区块内的第一笔交易,它负责将系统产生的奖励给挖出区块的矿工.由于它并不是普通意义上的转账,所以交易输入里并不需要引用任何一笔交易输出，下面就来创建铸币交易函数

`transaction.go`

```go
// 铸币交易
func NewCoinBaseTransaction(address string) *Transaction {
	txInput := &TXInput{[]byte{}, -1, "coinbase Data"}
	txOutput := &TXOuput{10, address}
	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOuput{txOutput}}
	//设置hash值
	txCoinbase.SetTxID()
	return txCoinbase
}

//设置交易的hash
func (tx *Transaction) SetTxID() {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	buffBytes := bytes.Join([][]byte{utils.IntToHex(time.Now().Unix()), buff.Bytes()}, []byte{})
	hash := sha256.Sum256(buffBytes)
	tx.TxID = hash[:]
}
```

接下来测试铸币交易是否被打包进区块中

`main.go`

```go
package main

import (
	"publicchain/cli"
)

func main() {

	//创建命令行工具
	cli := cli.CLI{}
	//激活cli
	cli.Run()
}

```

删除之前的数据库

```go
结果
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createblockchain -data "i am genesis block"
数据库不存在,创建创世区块：
52504: 000085576e27d2421c754d594191e2a56d5fb700e2bbd4344efef8b58c4a800e
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go printchain
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:000085576e27d2421c754d594191e2a56d5fb700e2bbd4344efef8b58c4a800e
        交易:
                交易ID:5338d886033a29c2e42ce7006918b787ca9b8017894a0e59a2c9ab1fdefece74
                Vins:
                        TxID:
                        Vout:-1
                        ScriptSiq:coinbase Data
                Vouts:
                        value:10
                        ScriptPubKey:i am genesis block
        时间:2022-07-11 11:27:42
        次数/;5250
```

可以看到，创建区块链生成的创世区块中，铸币交易已经被打包进去了，其实我们看到我们在创建区块链指定的-date数据被签名到交易的输出中了，其实实际情况是指定一个旷工的地址，然后把这个地址签名到输出中，这样就知道输出是该旷工的了

# 六、UTXO

区块中的数据往往以交易信息（Transaction）的形式存储。交易信息顾名思义，最初指的就是bitcoin中的各个用户的转账信息。这里提醒一下，随着区块链的发展，在非金融领域，人们还是习惯于将区块中储存的一条一条的有用数据称为交易信息。

既然在比特币中交易信息就是转账信息，我们不妨思考一下如何将“A把五块钱转给B”这个转账信息表示出来。也许你的表示如下：Sender：A；Reciever：B；Amount：5。很好，这很符合我们的直觉。在日常生活中要确认上述转账信息是否有效，只需要通过银行这个可信第三方机构就可以实现，因为银行记录了A与B各自的资产信息。现在回到区块链中，区块链作为一个去中心的分布式系统，其目的就是去掉可信第三方，此时我们如何确认这样一个转账信息有效了？聪明的你可能想到了和中本聪一样的办法，那就是在交易信息中向前回溯，找到以A作Reciever的前置交易信息，加和它们的Amout是否大于5，如果大于5本次转账就是有效的。

到这里，有的小朋友就有话要说了，如果我再进行“一次A把五块钱转给B”的转账，这次转账肯定也会被认为是有效的，一直重复，都将是有效的。很好，这个问题的关键所在就是在进行了交易回溯后，那些支持本次交易的前置交易信息没有被标记，导致这些前置交易信息被无限次的用于支持其它交易的回溯。我们需要做的就是在本次交易信息中标记出那些用于支持本次交易的前置交易信息。

可以看到，我们在确认转账信息进行回溯时，我们其实根本不关心前置交易信息的Sender是谁，我们只关心它们的Reciever和Amount，这就是就是比特币中强调的UTXO（Unspent Transaction Outputs）模型的基本思路，现在让我们看看比特币中的UTXO模型究竟是如何构建来实现上述功能的。（关于区块链为何使用UTXO我举了上述这个例子来讲解，可能一些同学还是不能理解，无妨，我们直接阅读并理解后文的代码来直观感受UTXO的精妙之处，这个东西有时候就是有点无法言传，这才需要看代码，talk is cheap, show me the code!）

## 1、UTXO结构

UTXO 代表 Unspent Transaction TxOutput,表示区块链上未经花费的交易输出。简单地说，UTXO还没有被包含在任何的交易输入中。根据UTXO可以知道对应TxOutput来自哪一笔交易，以及其在交易输出中的下标，这样就确定下来输出的来源，该输出没有花掉的话，那么组装一个UTXO对象来对应着它

在pbcc包下创建utxo.go

创建utxo的结构体

`utxo.go`

```go
//结构体UTXO，用于表示未花费的钱
type UTXO struct {
	TxID   []byte   //当前Transaction的交易ID
	Index  int      //这个在交易的输出的下标索引
	Output *TXOuput //输出
}
```

## 2、获取UTXO

有了UTXO的结构后，我们就可以创建获取未花费输出的方法，使其返回为UTXO类型的数组

其次，之前测试的都是单笔转账的交易，当出现多笔转账的交易时，我们现有的查询余额方法会不准确，为什么呢？

当一笔交易中有多个转账，当进行其中第二笔转账时，第一笔转账已经成功。但是，我们此时查询的依然是只是区块链上所有交易的UTXO，因此，我们还需要在UTXOs方法中加上当前未上链的所有交易的UTXO。

这时就有疑问了，不是只有上链的交易才会有效吗？事实是这样的，但是看目前的项目，由于还没有引入竞争挖矿的概念，每一次send必然会挖矿成功，其交易必然会上链，所以我们需要暂时这么做

在blockchain.go中添加UXTOs接口，用于找到某地址在所有交易中对应的的UTXOs

在这个寻找UTXO的过程中，怎么知道交易输出是未花费的呢？这和就需要去和该地址下已经花费的输出，已经被花费输出肯定是被当做某笔交易的输入了，把该地址的所有输入找到，然后和输出对比，就知道该输出是否被花了，如果没有被花掉，那么就组装一个UTXO

`blockchain.go`

```go
//找到某地址对应的所有UTXO
func (bc *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	/*
		1.先遍历未打包的交易(参数txs)，找出未花费的Output，为什么校验未打包的交易，教程有解释
		2.遍历数据库，获取每个块中的Transaction，找出未花费的Output。
	*/
	var unUTXOs []*UTXO                      //未花费输出
	spentTxOutputs := make(map[string][]int) //存储已经花费

	//添加先从txs遍历，查找未花费
	for i := len(txs) - 1; i >= 0; i-- {
		// 计算某笔交易中的输出是否已经被花费了
		unUTXOs = caculate(txs[i], address, spentTxOutputs, unUTXOs)
	}

	bcIterator := bc.Iterator()
	for {
		block := bcIterator.Next()
		//统计未花费
		//获取block中的每个Transaction
		for i := len(block.Txs) - 1; i >= 0; i-- {
			unUTXOs = caculate(block.Txs[i], address, spentTxOutputs, unUTXOs)
		}

		//结束迭代
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
	return unUTXOs
}

//找出这比交易中的属于对应地址的UTXO
func caculate(tx *Transaction, address string, spentTxOutputs map[string][]int, unUTXOs []*UTXO) []*UTXO {
	//判断是否为铸币交易，只有费铸币交易才有输入，先记录这个交易里面该地址的输入
	if !tx.IsCoinbaseTransaction() {
		//遍历交易里面的输入
		for _, in := range tx.Vins {
			//如果解锁 也就是这个输出是该地址花掉的了
			if in.UnLockWithAddress(address) {
				//回去交易的Hash
				key := hex.EncodeToString(in.TxID)
				// 记录已经花费的  记录交易的hash和索引
				spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
			}
		}
	}
outputs:
	// 遍历交易的输出
	for index, out := range tx.Vouts {
		// 如果解锁 那么这个输出是属于该地址
		if out.UnLockWithAddress(address) {
			//说明交易中有我的输出花费 需要进行对比
			if len(spentTxOutputs) != 0 {
				var isSpentUTXO bool
				// 遍历已经花费的输入
				for txID, indexArray := range spentTxOutputs {
					for _, i := range indexArray {
						//下标对上 交易的ID对上
						if i == index && txID == hex.EncodeToString(tx.TxID) {
							// 说明这个输出是被花掉的
							isSpentUTXO = true
							continue outputs //那么继续下一个
						}
					}
				}
				// 如果没花掉，那么加入utxo
				if !isSpentUTXO {
					// 组装一个UTXO对象
					utxo := &UTXO{tx.TxID, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}

			} else { //也就是该地址没有花费  则把属于该地址的输出都纳入UTXO
				utxo := &UTXO{tx.TxID, index, out}
				unUTXOs = append(unUTXOs, utxo)
			}
		}
	}
	return unUTXOs
}
```

## 3、转账交易的理解

上面已经实现了如何在未打包的交易中找到UTXO，到这里可能大家还会疑惑为什么会有UTXO，或者说UTXO怎么用

说到转账,就离不开交易.这里的转账便是普通交易,之前我们只实现了创币交易.这里需要实现普通交易.

为了更好地理解转账的过程,我们先将复杂问题简单化.假设每一个区块只有一笔交易,我们看一个简单的小🌰.

1.节点A挖到一个区块，产生25BTC的创币交易。由于是创币交易，其本身是不需要引用任何交易输出的，所以在输入对象TXInput的交易哈希为空，vount所在的下标为-1，数字签名为空或者随便填写；输出对象里btc拥有者为A，面值为25btc 创世区块交易结构

```
 txInput0 = &TXInput{[]byte{},-1,"Gensis Block"}
 txOutput0 = &TXOutput{25, "A"}  //在gaVouts索引为0

 CoinbaseTransaction{"00000",
			[]*TXInput{txInput0},
			[]*TXOutput{txOutput0}
}
```

2.A获得25btc后，他的好友B知道后向他索要10btc，大方的A便把10btc转给，此时 交易的输入为A上笔交易获得的btc，TXInput对象的交易ID为奖励chaors的上一个交易ID，vount下标为A的TXOutput下标，签名此时且认为是来自A，填作"A" 此时A的25btc面值的TXOutput就被花费不复存在了，那么A还应该有15btc的找零哪去了？系统会为A的找零新生成一个面值15btc的TXOutput，所以，这次有一个输入，两个输出。

> A(25) 给 B 转 10 -- >> A(15) + B(10)

这次的交易结构为:

```
 //输入
 txInput1 = &TXInput{"00000",0,"A"}
 //"00000" 相当于来自于哈希为"00000"的交易
 //索引为零，相当于上一次的txOutput0为输入

 //输出
 txOutput1 = &TXOutput{10, "B"}		//在该笔交易Vouts索引为0  A转给B的10btc产生的输出
 txOutput2 = &TXOutput{15, "A"}    //在该笔交易Vouts索引为1  给B转账产生的找零
 
 //对应这笔交易
 Transaction1{"11111"，
			[]*TXInput{txInput1}
			[]*TXOutput{txOutput1, txOutput2}
}
```

3.B感觉拥有比特币是一件很酷的事情，又来跟A要。出于兄弟情谊，A又转给B 7个BTC这次的交易结构为:

```
//输入
 txInput2 = &TXInput{"11111",1,"A"}

 //输出
 txOutput3 = &TXOutput{7, "B"}		  //在该笔交易Vouts索引为0
 txOutput4 = &TXOutput{8, "A"}   //在该笔交易Vouts索引为A
 
 //对应这笔交易
 Transaction2{"22222"，
			[]*TXInput{txInput2}
			[]*TXOutput{txOutput3, txOutput4}
}
```

4.消息传到他们共同的朋友C那里，C觉得btc很好玩向B索要15btc，B一向害怕C，于是尽管不愿意也只能屈服。

我们来看看B此时的所有财产:

```
txOutput1 = &TXOutput{10, "ww"}		//来自Transaction1(hash:11111)Vouts索引为0的输出   
txOutput3 = &TXOutput{7, "ww"}		//来自Transaction2(hash:2222)Vouts索引为0的输出
```

想要转账15btc,ww的哪一笔txOutput都不够，这个时候就需要用ww的两个txOutput都作为 输入,这次的交易结构为:

```
//输入：
txInput3 = &TXInput{"11111",1,"ww"}
txInput4 = &TXInput{"22222",3,"ww"}

//输出
 txOutput5 = &TXOutput{15, "C"}		//索引为0
 txOutput6 = &TXOutput{2, "B"}     //索引为1

//交易
 Transaction3{"33333"，
			[]*TXInput{txInput3, txInput4}
			[]*TXOutput{txOutput5, txOutput6}
}
```

现在,我们来总结一下上述几个交易.

> A
>
> > 1.从CoinbaseTransaction获得TXOutput0总额25 
> >
> > 2.Transaction1转给B10btc,TXOutput0被消耗,获得txOutput2找零15btc 
> >
> > 3.Transaction2转给B7Btc,txOutput2被消耗,获得txOutput4找零8btc 
> >
> > 4.最后只剩8btc的txOutput4作为未花费输出

> B
>
> > 1.从Transaction1获得TXOutput1,总额10btc 
> >
> > 2.从Transaction2获得TXOutput3,总额7btc 
> >
> > 3.Transaction3转给xyz15btc,TXOutput1和TXOutput3都被消耗,获得txOutput6找零2btc 
> >
> > 4.最后只剩2btc的txOutput6作为未花费输出

> C
>
> > 1.从Transaction3获得TXOutput5,总额15btc 
> >
> > 2.拥有15btc的TXOutput5作为未花费输出

经过这个例子,我们可以发现转账具备几个特点:

1.每笔转账必须有输入TXInput和输出TXOutput

2.每笔输入必须有源可查(TXInput.TxHash)

3.每笔输入的输出引用必须是未花费的(没有被之前的交易输入所引用)

4.TXOutput是一个不可分割的整体,一旦被消耗就不可用.消费额度不对等时会有找零(产生新的TXOutput)

这个🌰很重要,对于后面转账的代码逻辑是个扎实的基础准备.

接下来就要完成交易中的设置

在transaction.go下添加判断交易是普通交易还是铸币交易

`transaction.go`

```go
//判断当前交易是否是Coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxID) == 0 && tx.Vins[0].Vout == -1
}
```

TXInput和TXOutput解锁，也就是验证是不是自己花掉的和自己拥有的

上面unUTXOs方法求得是某一个address的所有UTXO，目前我们还没有引入钱包地址的概念，姑且理解这个address为用户名。我们要想保证查询的是某个用户(address)交易输入和输出是属于这个用户的，必须有一个保障的机制

分别在TXInput.go和TXOutput.go加入解锁

`TXInput.go`

```go
//判断当前txInput消费，和指定的address是否一致
func (txInput *TXInput) UnLockWithAddress(address string) bool {
	return txInput.ScriptSiq == address
}
```

`TXOutput.go`

```go
func (txOutput *TXOuput) UnLockWithAddress(address string) bool {
	return txOutput.ScriptPubKey == address
}
```

到这里UTXO的基础设置已经完成了，重点是理解代码中的原理，等下一个章节转账完成以后一起测试

# 七、转账

## 1、FindSpendableUTXOs

当我们理解了UTXO以后，其实转账交易就是去找到该地址的UTXO然后作为输入

当我们进行一笔转账时，交易输入有可能引用一个UTXO，也可能引用多个UTXO。在获取转账方所有的UTXO后，还需要找到符合条件的UTXO组合作为交易输入的引用。这个时候可能出现用户余额不足以转账的情况，也可能出现UTXO组合价值大于转账金额产生找零的情况

为了方便地判断UTXO来源以及计算转账后的找零，我们需要想办法在当前用户的所有UTXO中找到一个满足当前转账情况的UTXO集，并返回其UTXO总额和对应的UTXO集。而这个UTXO集是一个字典类型，键是UTXO来源交易的哈希，值对该交易下UTXO对应TXOutput在Vounts中的下标

提供一个寻找地址UTXO的升级接口

`blockchain.go`

```go
//查找UTXO的升级版
// 找出来以后判断余额够不够 然后把可花费的UTXO改成map[交易ID]输出的下标
func (bc *BlockChain) FindSpendableUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	var balance int64
	// 找出所有的utxo
	utxos := bc.UnUTXOs(from, txs)
	// 定义一个map 对应可花费的utxo
	spendableUTXO := make(map[string][]int)
	for _, utxo := range utxos {
		balance += utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxID)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		// 钱够了就不找了
		if balance >= amount {
			break
		}
	}
	// 要不就是余额不够转账的
	if balance < amount {
		fmt.Printf("%s 余额不足，总额：%d,需要：%d\n", from, balance, amount)
		os.Exit(1)
	}
	return balance, spendableUTXO
}
```

## 2、添加转账

有了上面的基础方法就可以对普通交易的构造做一个代码实现

添加普通转账的函数

`transaction.go`

```go
// 创建普通交易
func NewSimpleTransaction(from, to string, amount int64, bc *BlockChain, txs []*Transaction) *Transaction {
	var txInputs []*TXInput
	var txOutputs []*TXOuput
	// 未打包的中找到够花的utxo
	balance, spendableUTXO := bc.FindSpendableUTXOs(from, amount, txs)
	// 遍历spendableUTXO来组装
	for txID, indexArray := range spendableUTXO {
		txIDBytes, _ := hex.DecodeString(txID)
		for _, index := range indexArray {
			txInput := &TXInput{txIDBytes, index, from}
			txInputs = append(txInputs, txInput)
		}
	}
	//转账
	txOutput1 := &TXOuput{amount, to}
	txOutputs = append(txOutputs, txOutput1)
	// 找零
	txOutput2 := &TXOuput{balance - amount, from}
	txOutputs = append(txOutputs, txOutput2)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.SetTxID()
	return tx
}

```

理论上我们的交易是支持多笔转账的，可是上面构建交易的方法是针对一笔交易。所以，我们需要在发起交易挖掘区块的方法里对cli输入的多笔交易信息做一个遍历并生成多笔交易数据

同时，我们设计的BTC没有多节点之间的挖矿争夺，只要产生交易就会触发挖矿，所以我们提供一个挖矿的接口，等有交易的时候就调用，然后把交易的信息打包进区块中

`blockchain.go`

```go
//挖掘新的区块 有交易的时候就会调用
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	//新建交易
	//新建区块
	//将区块存入到数据库
	var txs []*Transaction
	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		// 新建一个普通交易
		tx := NewSimpleTransaction(from[i], to[i], amountInt, bc, txs)
		txs = append(txs, tx)
	}

	var block *Block    //数据库中的最后一个block
	var newBlock *Block //要创建的新的block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) //数据库中的最后一个block
		}
		return nil
	})
	newBlock = NewBlock(txs, block.Hash, block.Height+1)
	// 更新数据库
	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
}
```

接下来提供一个查询地址余额的函数

```go
// 获取余额
func (bc *BlockChain) GetBalance(address string, txs []*Transaction) int64 {
	unUTXOs := bc.UnUTXOs(address, txs)
	var amount int64
	for _, utxo := range unUTXOs {
		amount = amount + utxo.Output.Value
	}
	return amount
}
```

## 3、设置转账命令行参数

我们知道挖矿的目的是找到一个公认的记账人把当前的所有交易打包到区块并添加到区块链上.之前我们使用addBlock命令实现添加区块到区块链的,这里转账包含挖矿并添加到区块链.所以,我们需要在cli工具类里用转账命令send代替addBlock命令.

其次我们都知道,一次区块可以包括多个交易.因此,这里我们的转账命令要设计成支持多笔转账

命令行参数添加转账

命令行输入的都是字符串,要想让转账命令支持多笔转账,则输入的信息是json形式的数组.在编码实现解析并转账的时候,我们需要将Json字符串转化为数组类型，这个功能在utils里实现.

我们一般输入的转账命令是这样的:

```
send -from '["A", "B"]' -to '["B", "C"]' -amount '["5", "100"]'
```

> send 转账命令 from 发送方 to 接收方 amount 转账金额 三个参数的数组分别一一对应,上述命令表示: A转给B共5btc; B转给C的100btc.

在uitls下添加标准的JSON字符串转数组的接口方法

`utils.go`

```go
//Json字符串转为[] string数组
func JSONToArray(jsonString string) []string {
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}
	return sArr
}
```

提供转账命令的处理

`cli_send.go`

```go
//转账
func (cli *CLI) send(from, to, amount []string) {
	blockchain := pbcc.GetBlockchainObject()
	if blockchain == nil {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}
	blockchain.MineNewBlock(from, to, amount)
	defer blockchain.DB.Close()
}
```

提供余额查询命令的处理

`cli_getBalance.go`

```go
//查询余额
func (cli *CLI) getBalance(address string) {
	fmt.Println("查询余额：", address)
	bc := pbcc.GetBlockchainObject()
	if bc == nil {
		fmt.Println("数据库不存在，无法查询")
		os.Exit(1)
	}
	defer bc.DB.Close()
	balance := bc.GetBalance(address, []*pbcc.Transaction{})
	fmt.Printf("%s,一共有%d个Token\n", address, balance)
}
```

更新cli.go

`cli.go`

```go
package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"publicchain/utils"
)

//CLI结构体
type CLI struct {
}

//添加Run方法
func (cli *CLI) Run() {
	//判断命令行参数的长度
	isValidArgs()

	//创建flagset标签对象
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	//设置标签后的参数
	flagFromData := sendBlockCmd.String("from", "", "转帐源地址")
	flagToData := sendBlockCmd.String("to", "", "转帐目标地址")
	flagAmountData := sendBlockCmd.String("amount", "", "转帐金额")
	flagCreateBlockChainData := createBlockChainCmd.String("address", "", "创世区块交易地址")
	flagGetBalanceData := getBalanceCmd.String("address", "", "要查询的某个账户的余额")

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

	default:
		printUsage()
		os.Exit(1) //退出
	}

	if sendBlockCmd.Parsed() {
		if *flagFromData == "" || *flagToData == "" || *flagAmountData == "" {
			printUsage()
			os.Exit(1)
		}
		fmt.Println(*flagFromData)
		fmt.Println(*flagToData)
		fmt.Println(*flagAmountData)
		from := utils.JSONToArray(*flagFromData)
		to := utils.JSONToArray(*flagToData)
		amount := utils.JSONToArray(*flagAmountData)

		cli.send(from, to, amount)
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
	if getBalanceCmd.Parsed() {
		if *flagGetBalanceData == "" {
			fmt.Println("查询地址不能为空")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceData)
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
	fmt.Println("\tcreateblockchain -address DATA -- 创建创世区块")
	fmt.Println("\tsend -from From -to To -amount Amount - 交易数据")
	fmt.Println("\tprintchain - 输出信息")
	fmt.Println("\tgetbalance -address DATA -- 查询账户余额")
}
```

下面进行转账的测试

`main.go`

```go
package main

import (
	"publicchain/cli"
)

func main() {

	//创建命令行工具
	cli := cli.CLI{}
	//激活cli
	cli.Run()
}
```

删除数据库，重新创建区块链

```
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go 
Usage:
        createblockchain -address DATA -- 创建创世区块
        send -from From -to To -amount Amount - 交易数据
        printchain - 输出信息
        getbalance -address DATA -- 查询账户余额
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createblockchain -address "wp"
数据库不存在,创建创世区块：
138640: 00002b98d04c86044617908ef07882970f3d95d20c030413b9885d9cbee45da2
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go getbalance -address "wp"
查询余额： wp
wp,一共有10个Token
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["wp"]' -to '["zt"]' -amount 
40646: 0000c48ecbce275cbcac4e83108615be665a06094481608237f37f16c83bef55
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["wp","zt"]' -to '["mimi","mimi"]' -amount '["5","5"]' 
58496: 000061d4db07a716ea7db09371a9470a9053324850d857ae1b534c9481cf6bca
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go getbalance -address "mimi"
查询余额： mimi
mimi,一共有10个Token
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["wp"]' -to '["zt"]' -amount '["5"]'              
wp 余额不足，总额：0,需要：5
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go printchain
第3个区块的信息:
        高度:2
        上一个区块的hash:0000c48ecbce275cbcac4e83108615be665a06094481608237f37f16c83bef55
        当前的hash:000061d4db07a716ea7db09371a9470a9053324850d857ae1b534c9481cf6bca
        交易:
                交易ID:fa5324f618678b0f29568976e2efd26491bcd3975dd276148350ce24eda1abb8
                Vins:
                        TxID:8d41ef31e6fcbd18be0cc764a931f79cb6c79c6f735786a3d59c4cbb0ac5b67b
                        Vout:1
                        ScriptSiq:wp
                Vouts:
                        value:5
                        ScriptPubKey:mimi
                        value:0
                        ScriptPubKey:wp
                交易ID:149216c806ee26d389b5b9db6ded540de75ac95cf8fb0548110ea5d85cfa5d75
                Vins:
                        TxID:8d41ef31e6fcbd18be0cc764a931f79cb6c79c6f735786a3d59c4cbb0ac5b67b
                        Vout:0
                        ScriptSiq:zt
                Vouts:
                        value:5
                        ScriptPubKey:mimi
                        value:0
                        ScriptPubKey:zt
        时间:2022-07-11 14:14:48
        次数/;58496
第2个区块的信息:
        高度:1
        上一个区块的hash:00002b98d04c86044617908ef07882970f3d95d20c030413b9885d9cbee45da2
        当前的hash:0000c48ecbce275cbcac4e83108615be665a06094481608237f37f16c83bef55
        交易:
                交易ID:8d41ef31e6fcbd18be0cc764a931f79cb6c79c6f735786a3d59c4cbb0ac5b67b
                Vins:
                        TxID:577647ab0ec0be197a68c07505f3fb244b216b0c546b442d477cf57d02ff1954
                        Vout:0
                        ScriptSiq:wp
                Vouts:
                        value:5
                        ScriptPubKey:zt
                        value:5
                        ScriptPubKey:wp
        时间:2022-07-11 14:13:09
        次数/;40646
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:00002b98d04c86044617908ef07882970f3d95d20c030413b9885d9cbee45da2
        交易:
                交易ID:577647ab0ec0be197a68c07505f3fb244b216b0c546b442d477cf57d02ff1954
                Vins:
                        TxID:
                        Vout:-1
                        ScriptSiq:coinbase Data
                Vouts:
                        value:10
                        ScriptPubKey:wp
        时间:2022-07-11 14:11:29
        次数/;138640
```

可以看见现在已经可以正常转账了

同时我们发余额不够的话是无法进行转账的

# 八、钱包地址

区块链中的钱包与现实生活中的钱包有很大不同，这里我说说本人自己的理解。可以认为区块链中的一个钱包就对应了一个区块链用户，用户的密钥对都由钱包保存。现实中的钱包保存的资产就是真实的现金，在需要做交易时直接划拨现金即可；而在区块链中的钱包保存的资产不是现金（真实的货币），而是密钥对所指向的区块链中的相应UTXO总额，划拨资产要先通过密钥对证明这些UTXO的所属权。换句话说，区块链中的钱包根本不存储资产（UTXO自始至终都保存于区块链中），它只是使用密钥对帮助管理用户的个人资产。

在前面的章节中我们经常提到一个概念，那就是地址。区块链中的地址作用只有一个，那就是唯一指向一个用户，使得UTXO可以通过指向该地址来流向该用户。我们现阶段的send命令是用的昵称来作为地址，一个昵称指向一个特定的用户，但是这种地址表示方式终究很简陋，首先昵称太容易重复，其次没有一种手段来证明昵称的所属权（如何证明你的昵称是你的）。结合非对称密钥，我们知道公钥是可以公开给任何人的，同时也能够作为身份标识，用户通过掌握私钥也能够证明公钥的所属权，那为什么不使用公钥来作为地址了。

在最初的比特币中，的确是使用公钥来作为地址的，也即构造的交易信息是从公钥指向公钥。而在后续的版本更迭中，逐渐不再直接使用公钥作为地址，而是使用公钥进行一些列哈希操作得到的值作为地址，这是因为公钥哈希值能够在指向原用户个体的同时提升匿名性。

以椭圆曲线加密算法为例，公钥哈希值的生成方法如下：

[![./goblockchain5/pic1.png](https://www.krad.top/goblockchain05/goblockchain5/pic1.png)](https://www.krad.top/goblockchain05/goblockchain5/pic1.png)

可以看到，公钥哈希就是将公钥连续做了两次哈希操作得到（一次sha256一次ripemd160）。上图在公钥哈希的基础上还生成了钱包地址，钱包地址其实就是公钥哈希增加一个版本号位与四个检查位生成，最后转为比特币专门的Base58编码输出。

总结一下，公钥可以推得公钥哈希，但是公钥哈希不能推回公钥，同时公钥哈希和钱包地址可以通过Base58互相转化

之前我们的项目中转账什么的都是使用的字符串做用户名，但是在比特币种并没有用户账户的概念。所有的交易都是基于地址进行转账的，所谓的地址本质是一个公钥，地址只是把公钥通过一个转化用人们可读的方式表现出来

地址是基于一系列加密算法完成的，关于加密算法这里也不再叙述，重点看BTC系统里面是怎么实现的钱包的

关于地址和加密算法知识这里就不叙述了

## 1、钱包

就像我们生活中我们的纸币往往会放在自己的钱包里，比特币也是如此，每一个钱包有一个地址作为唯一标示。而这个地址是由公钥经过几次哈希算法再经过Base58编码转化而成的。

公钥是由私钥产生的，公私钥总是成对出现的。公钥不是敏感信息，可以告诉其他人。从本质上讲，我们安装一个比特币钱包应用，其实是从比特币客户端生成一个公私钥对，私钥代表你对该钱包的控制权，公钥代表该钱包的地址

新建wallet包，在包下创建wallet.go

创建钱包的结构体，包括私钥和公钥两个字段

`wallet.go`

```go
//单个钱包地址结构
type Wallet struct {
	PrivateKey ecdsa.PrivateKey //私钥
	PublicKey  []byte           //公钥
}
```

要创建一个钱包地址，首先要获得一个私钥，那么私钥的利用椭圆曲线算法生成，然后用私钥生成公钥

```go
//产生一对密钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	/*
		1.通过椭圆曲线算法，随机产生私钥
		2.根据私钥生成公钥

		elliptic:椭圆
		curve：曲线
		ecc：椭圆曲线加密
		ecdsa：elliptic curve  digital signature algorithm，椭圆曲线数字签名算法
			比特币使用SECP256K1算法，p256是ecdsa算法中的一种
	*/
	//椭圆加密
	curve := elliptic.P256() //椭圆加密算法，得到一个椭圆曲线值，全称：SECP256k1
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	//生成公钥
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}
```

提供一个创建新钱包地址的接口

```go
//获取一个钱包
func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()
	return &Wallet{privateKey, publicKey}
}
```

我们来看一下生成的公钥和私钥目前是什么样的格式

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/wallet"
)

func main() {
	wallet := wallet.NewWallet()
	fmt.Println(wallet)
}
```

```go
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go 
&{{{{0xc00002c0c0} 113774222260799301769470755875759489551211644360925663930209383985928336308010 63004116631675219156289570543642389998102637283276952709226016212785646797383} 41670407505949749715512832450925072617104817422901570266041573403244642643877} [251 137 237 129 254 141 102 28 231 7 239 8 192 15 194 234 22 121 135 185 144 16 242 106 215 111 124 105 130 31 247 42 139 75 16 249 73 210 222 88 181 166 91 140 193 70 110 102 32 235 57 181 56 60 132 80 90 44 83 156 64 180 2 71]}
```

当前生成的公钥和私钥的形式都不是我们认识的样子，接下来还要继续转化

## 2、生成address

从公钥得到一个公钥哈希需要五步走：

1.公钥经过两次哈希(SHA256+RIPEMD160)得到一个字节数组PubKeyHash

2.PubKeyHash+交易版本Version拼接成一个新的字节数组Version_PubKeyHash

3.对Version_PubKeyHash进行两次哈希(SHA256)并按照一定规则生成校验和CheckSum

4.Version_PubKeyHash+CheckSum拼接成Version_PubKeyHash_CheckSum字节数组，也就是公钥的hash

5.对Version_PubKeyHash_CheckSum进行Base58编码即可得到地址Address

通过上面五个步骤，我们知道Address讲过Base58解码后由三部分组成

首先在配置信息里面添加版本和校验和的设置

`config.go`

```go
const Version = byte(0x00)   //版本
const AddressChecksumLen = 4 //校验和的长度
```

还要准备一下base58的编码，创建一个crypto包，包下创建base58.go，提供一个利用base58加密和解密的方法

`base58.go`

```go
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

//字节数组转Base58，编码
func Base58Encode(input []byte) []byte {
	var result []byte
	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}
	utils.ReverseBytes(result)
	for b := range input {
		if b == 0x00 {
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return result

}

//Base58转字节数组，解码
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0
	for b := range input {
		if b == 0x00 {
			zeroBytes++
		}
	}
	payload := input[zeroBytes:]
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}
```

在utils.go下提供一个字节数组翻转的接口

`utils.go`

```go
//字节数组反转
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
```

接下来我们可以利用公钥进行进一步的操作了，在wallet.go添加生成地址方法

`wallet.go`

```go
//根据一个公钥获取对应的地址
func (w *Wallet) GetAddress() []byte {
	//先将公钥进行一次hash256，一次160,得到pubKeyHash
	pubKeyHash := PubKeyHash(w.PublicKey)
	//添加版本号
	versioned_payload := append([]byte{conf.Version}, pubKeyHash...)
	// 获取校验和，将pubKeyhash，两次sha256后，取前4位
	checkSumBytes := CheckSum(versioned_payload)
	// 获得公钥的hash
	full_payload := append(versioned_payload, checkSumBytes...)
	//Base58
	address := crypto.Base58Encode(full_payload)
	return address

}

//一次sha256,再一次ripemd160,得到publicKeyHash
func PubKeyHash(publicKey []byte) []byte {
	//对公钥进行一个sha256
	hasher := sha256.New()
	hasher.Write(publicKey)
	hash := hasher.Sum(nil)
	//对结果ripemd160
	ripemder := ripemd160.New()
	ripemder.Write(hash)
	pubKeyHash := ripemder.Sum(nil)
	return pubKeyHash
}

//获取验证码：将公钥哈希两次sha256,取前4位，就是校验和
func CheckSum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:conf.AddressChecksumLen]
}
```

到这里项目可能会依赖报错，由于引入了`golang.org/x/crypto/ripemd160`，我们只需要在终端运行`go mod tidy`即可

验证地址是否有效

`wallet.go`

```go
//判断地址是否有效
func IsValidForAddress(address []byte) bool {
	full_payload := crypto.Base58Decode(address)
	checkSumBytes := full_payload[len(full_payload)-conf.AddressChecksumLen:]
	versioned_payload := full_payload[:len(full_payload)-conf.AddressChecksumLen]
	checkBytes := CheckSum(versioned_payload)
	return bytes.Equal(checkSumBytes, checkBytes)
}
```

接下来我们测试一下地址的获取

`main.go`

```go
package main

import (
	"fmt"
	"publicchain/conf"
	"publicchain/crypto"
	"publicchain/wallet"
)

func main() {
	w := wallet.NewWallet()
	fmt.Println(w.PrivateKey)
	fmt.Println("-------------------------")
	fmt.Println(w.PublicKey)
	w1 := wallet.PubKeyHash(w.PublicKey)
	fmt.Println("-------------------------")
	fmt.Println(w1)
	//添加版本号
	versioned_payload := append([]byte{conf.Version}, w1...)
	fmt.Println("-------------------------")
	fmt.Println(versioned_payload)
	c := wallet.CheckSum(w1)
	fmt.Println("-------------------------")
	fmt.Println(c)
	full_payload := append(versioned_payload, c...)
	fmt.Println("-------------------------")
	fmt.Println(full_payload)
	address := crypto.Base58Encode(full_payload)
	fmt.Println("-------------------------")
	fmt.Println(address)
	fmt.Println("-------------------------")
	fmt.Println(string(address))
}
```

```
{{{0xc00002d5c0} 109203541070387670543739522498062103835221283865416288315100845438904651605126 83421243691636626849278778179818700726006300561608206201274488021064341196133} 9250984120538218849026813761556377864057645595489127958144276732429824092898}
-------------------------
[241 111 3 142 38 186 136 253 44 231 189 200 163 226 244 95 220 3 162 38 111 42 164 72 188 27 50 9 8 253 228 134 184 110 192 45 253 121 11 210 113 234 109 193 30 152 177 11 195 186 149 77 224 66 126 193 55 171 8 132 73 159 13 101]
-------------------------
[136 246 74 14 31 192 105 139 31 16 48 235 34 227 233 13 243 187 14 132]
-------------------------
[0 136 246 74 14 31 192 105 139 31 16 48 235 34 227 233 13 243 187 14 132]
-------------------------
[202 122 191 243]
-------------------------
[0 136 246 74 14 31 192 105 139 31 16 48 235 34 227 233 13 243 187 14 132 202 122 191 243]
-------------------------
[49 68 86 66 119 90 119 102 113 87 110 72 55 56 49 71 113 65 74 102 106 74 51 116 120 112 116 111 66 112 74 56 53 120]
-------------------------
1DVBwZwfqWnH781GqAJfjJ3txptoBpJ85x
```

这里我们就很直观的能够看到一个地址生成的全部过程了，最后把[]byte转成string发现获取的地址已经是我们认识的模样了

## 3、钱包集合

有了钱包后，链上所有钱包需要一个统一的管理，这就有赖于Wallets类。当我们在比特币上创建一个钱包的地址后，那么这个钱包的地址将一直存在。因此，Wallets钱包集也需要实现数据持久化，这里我们采用文件去存储钱包集

在配置文件中设置钱包集合的名字

`config.go`

```go
const WalletFile = "Wallets.dat" //钱包集
```

创建wallets.go，钱包集合里面，使用钱包地址映射到钱包结构

`wallets.go`

```go
//钱包集
type Wallets struct {
	WalletsMap map[string]*Wallet
}

// 获取钱包集，如果数据库有就从数据库获取，如果没有就创建
func NewWallets() *Wallets {
	//判断钱包文件是否存在
	if _, err := os.Stat(conf.WalletFile); os.IsNotExist(err) {
		fmt.Println("文件不存在")
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets
	}
	//否则读取文件中的数据
	fileContent, err := ioutil.ReadFile(conf.WalletFile)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	return &wallets
}
```

有了钱包集合了以后，那么就可以创建钱包了，同时创建一个钱包一个就会将钱包存入本地一下

`wallets.go`

```go
//钱包集创建一个新钱包
func (ws *Wallets) CreateNewWallet() {
	wallet := NewWallet()
	fmt.Printf("创建钱包地址：%s\n", wallet.GetAddress())
	ws.WalletsMap[string(wallet.GetAddress())] = wallet
	//将钱包保存
	ws.SaveWallets()
}

/*
要让数据对象能在网络上传输或存储，我们需要进行编码和解码。
现在比较流行的编码方式有JSON,XML等。然而，Go在gob包中为我们提供了另一种方式，该方式编解码效率高于JSON。
gob是Golang包自带的一个数据结构序列化的编码/解码工具
*/
func (ws *Wallets) SaveWallets() {
	var content bytes.Buffer
	//注册的目的，为了可以序列化任何类型，wallet结构体中有接口类型。将接口进行注册
	gob.Register(elliptic.P256()) //gob是Golang包自带的一个数据结构序列化的编码/解码工具
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	//将序列化后的数据写入到文件，原来的文件中的内容会被覆盖掉
	err = ioutil.WriteFile(conf.WalletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
```

下面我们测试一下钱包集的使用

`main.go`

```go
package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"publicchain/conf"
	"publicchain/wallet"
)

func main() {
	wallets := wallet.NewWallets()
	wallets.CreateNewWallet()
	fileContent, err := ioutil.ReadFile(conf.WalletFile)
	if err != nil {
		log.Panic(err)
	}
	var ws wallet.Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&ws)
	if err != nil {
		log.Panic(err)
	}
	for address, _ := range ws.WalletsMap {
		fmt.Println(address)
	}
}
```

创建完钱包地址以后输出钱包集看下

```
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go
文件不存在
创建钱包地址：1KL1V2xDtgj1ATTBfRNPxtd6SxYxPEQhN3
1KL1V2xDtgj1ATTBfRNPxtd6SxYxPEQhN3
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go
创建钱包地址：1NfF1ThnrWKyQT8DndCEPDLp3kdQ1mjUqt
1KL1V2xDtgj1ATTBfRNPxtd6SxYxPEQhN3
1NfF1ThnrWKyQT8DndCEPDLp3kdQ1mjUqt
```

# 九、使用地址签名

## 1、交易集成Address

由于引入了地址和公私钥的概念，这里就可以给交易输入引入签名和公钥属性。这里且不论什么是签名，公钥代表这笔输入属于哪一个钱包

对交易的输入进行修改

`TXInput.go`

```go
//输入结构体
type TXInput struct {
	TxID      []byte //交易的ID
	Vout      int    //存储Txoutput的vout里面的索引
	Signature []byte //数字签名
	PublicKey []byte //公钥
}
```

之前我们验证交易输入是否属于一个账户时，由于我们设定的账户值和公钥直接是一个字符串形式的用户名，直接比较即可。现在输入的是一个地址，又该怎么办呢？

我们知道地址是由公钥进行多次哈希和按一定规则运算得出的，由于哈希是不可逆的，我们不可能根据地址反推出公钥然后和交易输入的公钥属性去作比较。这个时候，就需要拿地址Base58解码后得到的公钥哈希去和拿按交易输入的公钥进行两次哈希得到的值进行比较即可

`TXInput.go`

```go
//判断当前txInput消费，和指定的address是否一致
func (txInput *TXInput) UnLockWithAddress(pubKeyHash []byte) bool {
	//把交易里面的公钥进行计算
	publicKey := wallet.PubKeyHash(txInput.PublicKey)
	return bytes.Equal(pubKeyHash, publicKey)
}
```

我们知道比特币种对每一笔交易中的输出都会做一个锁定，将其锁定为某一个钱包所拥有；当这笔交易输出用于下一笔交易作为交易输入时需要进行解锁操作以保障花费的这笔TXOutput是属于当前转账方钱包的。

引入钱包的概念后，我们就可以实现TXOutput的锁定和解锁

`TXOutput.go`

```go
//输出结构体
type TXOuput struct {
	Value      int64  //面值
	PubKeyHash []byte // 公钥
}

//判断当前txOutput消费，和指定的address是否一致
func (txOutput *TXOuput) UnLockWithAddress(address string) bool {
	fullPayloadHash := crypto.Base58Decode([]byte(address))
	pubKeyHash := fullPayloadHash[1 : len(fullPayloadHash)-4]
	return bytes.Equal(txOutput.PubKeyHash, pubKeyHash)
}

// 创建新的输出
func NewTXOuput(value int64, address string) *TXOuput {
	txOutput := &TXOuput{value, nil}
	txOutput.Lock(address)
	return txOutput
}
// 把输出锁定给某地址
func (txOutput *TXOuput) Lock(address string) {
	publicKeyHash := crypto.Base58Decode([]byte(address))
	txOutput.PubKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]
}
```

现在输入和输出的结构发生了变化，那么之前交易的一些关于输入输出的地方都要修改一下，把有报错提示的地方关于输入输出的都按现在的结构来修改一下

打印区块链

铸币交易的输入和输出

普通交易的输入和输出

## 2、签名

通俗地讲，数字签名就是每一笔交易的证明。如果一个交易的数字签名是无效的，那么这笔交易就会被认为是无效的，因此，这笔交易也就无法被加到区块链中

数字签名的主要作用有两点：其一，证明该交易是转账方发起的。其二，证明交易信息没有被更改

当某个地址发起一笔转账时，需要首先打包成交易并对该交易进行数字摘要组成一段字符串，然后再用自己的私钥对摘要字符串进行加密形成数字签名。发起转账的用户会把交易信息和数字签名一起发送给矿工，矿工会用转账用户的公钥对数字签名进行验签，如果验证成功说明交易确实是该转账方发起的且交易信息未被篡改，交易有效可以打包到区块内

从前面的学习中，我们知道一笔交易包含交易哈希可以证明整个交易信息是否被篡改；交易输入引用的之前交易产生的UTXO可以证明交易发起方是谁；因此，我们签名的内容包括交易的交易哈希和交易输入里引用的TXOutput的公钥哈希，也就是要对交易的hash和输入所引用的输出的公钥哈希进行签名

签名的产生依赖于EDSA椭圆曲线加密算法，通过私钥PrivateKey经过一定运算得到签名。

由于在签名过程中，我们的目的是得到交易的签名。因此为了保证交易其他信息不被改变，我们需要在计算过程中对交易进行一个拷贝。由于计算前签名为空，签名的过程并不需要交易输入的公钥值。因此，拷贝的交易的签名和公钥置为nil。

由于创币交易的特殊性(其没有交易输入)，所以创币交易不需要进行签名

交易的签名

`transaction.go`

```go
//对交易签名
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]*Transaction) {
	//如果时coinbase交易，无需签名
	if tx.IsCoinbaseTransaction() {
		return
	}
	//input没有对应的transaction,无法签名   就是说这个交易的输入是伪造的
	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxID)].TxID == nil {
			log.Panic("当前的input没有对应的transaction")
		}
	}

	//获取Transaction的部分数据的副本 把输入的签名和公钥都置为空
	txCopy := tx.TrimmedCopy()

	for index, input := range txCopy.Vins {
		//通过输入的交易ID字段找到之前的交易
		prevTx := prevTXs[hex.EncodeToString(input.TxID)]
		input.Signature = nil //双保险 再置空一次
		// 钱是谁的 找到源头 再赋予公钥
		input.PublicKey = prevTx.Vouts[input.Vout].PubKeyHash //设置input的公钥为对应输出的公钥哈希
		data := txCopy.getData()                              //设置新的txID
		input.PublicKey = nil                                 //再将publicKey置为nil  为什么这里还要重新置空

		//签名
		/*
			通过 privKey 对 txCopy.ID 进行签名。
			一个 ECDSA 签名就是一对数字，我们对这对数字连接起来，并存储在输入的 Signature 字段。
		*/
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, data)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vins[index].Signature = signature
	}
}

//获取签名所需要的Transaction的副本
//创建tx的副本：需要剪裁数据
/*
	TxID，
		[]*TxInput,
		TxInput中，去除sign，publicKey
		[]*TxOutput
		这个副本包含了所有的输入和输出，但是 TXInput.Signature 和 TXIput.PubKey 被设置为 nil。
*/
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TXInput
	var outputs []*TXOuput
	for _, input := range tx.Vins {
		inputs = append(inputs, &TXInput{input.TxID, input.Vout, nil, nil})
	}
	for _, output := range tx.Vouts {
		outputs = append(outputs, &TXOuput{output.Value, output.PubKeyHash})
	}
	txCopy := Transaction{tx.TxID, inputs, outputs}
	return txCopy
}

//把交易序列化成字节数组
func (tx *Transaction) Serialize() []byte {
	jsonByte, err := json.Marshal(tx)
	if err != nil {
		log.Panic(err)
	}
	return jsonByte
}

// 交易的hash置为空 然后重新获取交易的hash
func (tx Transaction) getData() []byte {
	txCopy := tx
	txCopy.TxID = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
```

既然有交易签名，那么在打包交易到区块时就需要对交易进行签名验证，交易验签的前半部分和交易签名，不同的是后部分验签通过对签名计算出一对数字，然后利用公钥通过EDSA构造出一个ecdsa.PublicKey，最后借助EDSA加密算法的验证功能去验证构造的公钥，签名对应的一堆数字以及交易哈希。验证通过，说明交易验签成功

```go
//验证数字签名
func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	//没有对应的transaction,无法签名
	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxID)].TxID == nil {
			log.Panic("当前的input没有对应的transaction,无法验证")
		}
	}
	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()
	for index, input := range tx.Vins {
		prevTx := prevTXs[hex.EncodeToString(input.TxID)]
		txCopy.Vins[index].Signature = nil
		txCopy.Vins[index].PublicKey = prevTx.Vouts[input.Vout].PubKeyHash
		data := txCopy.getData()
		txCopy.Vins[index].PublicKey = nil

		//签名中的s和r
		r := big.Int{}
		s := big.Int{}
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:sigLen/2])
		s.SetBytes(input.Signature[sigLen/2:])

		//通过公钥，产生新的s和r，与原来的进行对比
		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PublicKey)
		x.SetBytes(input.PublicKey[:keyLen/2])
		y.SetBytes(input.PublicKey[keyLen/2:])

		//根据椭圆曲线，以及x，y获取公钥
		//我们使用从输入提取的公钥创建了一个 ecdsa.PublicKey
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y} //
		//这里我们解包存储在 TXInput.Signature 和 TXInput.PubKey 中的值，
		// 因为一个签名就是一对数字，一个公钥就是一对坐标。
		// 我们之前为了存储将它们连接在一起，现在我们需要对它们进行解包在 crypto/ecdsa 函数中使用。

		//验证
		//在这里：我们使用从输入提取的公钥创建了一个 ecdsa.PublicKey，通过传入输入中提取的签名执行了 ecdsa.Verify。
		// 如果所有的输入都被验证，返回 true；如果有任何一个验证失败，返回 false.
		if !ecdsa.Verify(&rawPubKey, data, &r, &s) {
			//公钥，要验证的数据，签名的r，s
			return false
		}
	}
	return true
}
```

对交易签名首先需要某个方法能获取到某个交易

`blockchain.go`

```go
//根据交易ID查找对应的Transaction
func (bc *BlockChain) FindTransactionByTxID(txID []byte, txs []*Transaction) *Transaction {
	itertaor := bc.Iterator()
	//先遍历txs
	for _, tx := range txs {
		if bytes.Equal(txID, tx.TxID) {
			return tx
		}
	}

	for {
		block := itertaor.Next()
		for _, tx := range block.Txs {
			if bytes.Equal(txID, tx.TxID) {
				return tx
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return &Transaction{}
}
```

区块链层级的交易签名和验签函数

`blockchain.go`

```go
// 区块链层的交易签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTxs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTxs[hex.EncodeToString(prevTx.TxID)] = prevTx
	}

	tx.Sign(privKey, prevTxs)
}

//区块链层的交易签名验证
func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {
	prevTXs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTXs[hex.EncodeToString(prevTx.TxID)] = prevTx
	}
	return tx.Verify(prevTXs)
}
```

创建交易的时候也需要进行签名

`blockchain.go`

```go
// 创建普通交易
func NewSimpleTransaction(from, to string, amount int64, bc *BlockChain, txs []*Transaction) *Transaction {
	var txInputs []*TXInput
	var txOutputs []*TXOuput
	balance, spendableUTXO := bc.FindSpendableUTXOs(from, amount, txs)
	//获取钱包
	wallets := wallet.NewWallets()
	wallet := wallets.WalletsMap[from]
	for txID, indexArray := range spendableUTXO {
		txIDBytes, _ := hex.DecodeString(txID)
		for _, index := range indexArray {
			txInput := &TXInput{txIDBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}
	//转账
	txOutput1 := NewTXOuput(amount, to)
	txOutputs = append(txOutputs, txOutput1)
	//找零
	txOutput2 := NewTXOuput(balance-amount, from)
	txOutputs = append(txOutputs, txOutput2)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.SetTxID()
	//进行签名
	bc.SignTransaction(tx, wallet.PrivateKey, txs)
	return tx
}
```

那么在打包交易的挖矿中，就需要对交易进行校验

`blockchain.go`

```go
//挖掘新的区块 有交易的时候就会调用
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction
	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		tx := NewSimpleTransaction(from[i], to[i], amountInt, bc, txs)
		txs = append(txs, tx)
	}
	var block *Block    //数据库中的最后一个block
	var newBlock *Block //要创建的新的block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) //数据库中的最后一个block
		}
		return nil
	})

	//在建立新区块前，对txs进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if !bc.VerifyTransaction(tx, _txs) {
			log.Panic("签名验证失败")
		}
		_txs = append(_txs, tx)
	}

	newBlock = NewBlock(txs, block.Hash, block.Height+1)

	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})

}
```

到目前项目应该还有报错的地方，由于之前改动了输入输出的结构，但是其他的接口并没有改动

计算地址UTXO接口，caculate里面解锁输出的地方需要进行修改

`blockchain.go`

```go
func caculate(tx *Transaction, address string, spentTxOutputs map[string][]int, unUTXOs []*UTXO) []*UTXO {
	//先遍历TxInputs，表示花费
	if !tx.IsCoinbaseTransaction() {
		for _, in := range tx.Vins {
			//如果解锁
			fullPayloadHash := crypto.Base58Decode([]byte(address))
			pubKeyHash := fullPayloadHash[1 : len(fullPayloadHash)-conf.AddressChecksumLen]

			if in.UnLockWithAddress(pubKeyHash) {
				key := hex.EncodeToString(in.TxID)
				spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
			}
		}
	}
outputs:
	for index, out := range tx.Vouts {
		if out.UnLockWithAddress(address) {
			if len(spentTxOutputs) != 0 {
				var isSpentUTXO bool

				for txID, indexArray := range spentTxOutputs {
					for _, i := range indexArray {
						if i == index && txID == hex.EncodeToString(tx.TxID) {
							isSpentUTXO = true
							continue outputs
						}
					}
				}
				if !isSpentUTXO {
					utxo := &UTXO{tx.TxID, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}

			} else {
				utxo := &UTXO{tx.TxID, index, out}
				unUTXOs = append(unUTXOs, utxo)
			}
		}
	}
	return unUTXOs
}
```

输出区块链`Printchains`接口需要进行修改

`blockchain.go`

```go
func (bc *BlockChain) PrintChains() {
	//获取迭代器对象
	bcIterator := bc.Iterator()

	//循环迭代
	for {
		block := bcIterator.Next()
		fmt.Printf("第%d个区块的信息:\n", block.Height+1)
		//获取当前hash对应的数据，并进行反序列化
		fmt.Printf("\t高度:%d\n", block.Height)
		fmt.Printf("\t上一个区块的hash:%x\n", block.PrevBlockHash)
		fmt.Printf("\t当前的hash:%x\n", block.Hash)
		//fmt.Printf("\t数据：%v\n", block.Txs)
		fmt.Println("\t交易:")
		for _, tx := range block.Txs {
			fmt.Printf("\t\t交易ID:%x\n", tx.TxID)
			fmt.Println("\t\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\t\tTxID:%x\n", in.TxID)
				fmt.Printf("\t\t\tVout:%d\n", in.Vout)
				fmt.Printf("\t\t\tPublicKey:%v\n", in.PublicKey)
			}
			fmt.Println("\t\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\t\tvalue:%d\n", out.Value)
				fmt.Printf("\t\t\tPubKeyHash:%v\n", out.PubKeyHash)
			}
		}
		fmt.Printf("\t时间:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数:%d\n", block.Nonce)

		//3.直到父hash值为0
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}
```

铸币交易需要进行修改

`trancaction.go`

```go
func NewCoinBaseTransaction(address string) *Transaction {
	txInput := &TXInput{[]byte{}, -1, nil, []byte{}}
	txOutput := NewTXOuput(10, address)
	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOuput{txOutput}}
	txCoinbase.SetTxID()
	return txCoinbase
}
```

现在已经完成了代码更新，项目此时应该没有报错了，大家仔细核对

## 3、设置命令行参数

创建cli_createwallet.go

处理创建新钱包地址的命令

`cli_createwallet.go`

```go
// 创建一个新钱包地址
func (cli *CLI) createWallet() {
	wallets := wallet.NewWallets()
	wallets.CreateNewWallet()
}
```

创建cli_getaddresslist.go

处理获取地址列表的命令

`cli_getaddresslist.go`

```go
// 打印所有钱包地址
func (cli *CLI) addressLists() {
	fmt.Println("打印所有的钱包地址")
	//获取
	Wallets := wallet.NewWallets()
	for address := range Wallets.WalletsMap {
		fmt.Println("address:", address)
	}
}
```

更新cli.go

`cli.go`

```go
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

	//创建flagset标签对象
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	addressListsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	//设置标签后的参数
	flagFromData := sendBlockCmd.String("from", "", "转帐源地址")
	flagToData := sendBlockCmd.String("to", "", "转帐目标地址")
	flagAmountData := sendBlockCmd.String("amount", "", "转帐金额")
	flagCreateBlockChainData := createBlockChainCmd.String("address", "", "创世区块交易地址")
	flagGetBalanceData := getBalanceCmd.String("address", "", "要查询的某个账户的余额")

	//解析
	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
		//fmt.Println("----",os.Args[2:])

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
		//fmt.Println("====",os.Args[2:])

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
		cli.send(from, to, amount)
	}
	if printChainCmd.Parsed() {
		cli.printChains()
	}

	if createBlockChainCmd.Parsed() {
		if !wallet.IsValidForAddress([]byte(*flagCreateBlockChainData)) {
			fmt.Println("创建地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockChainData)
	}

	if getBalanceCmd.Parsed() {
		if !wallet.IsValidForAddress([]byte(*flagGetBalanceData)) {
			fmt.Println("查询地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceData)

	}

	if createWalletCmd.Parsed() {
		//创建钱包
		cli.createWallet()
	}

	//获取所有的钱包地址
	if addressListsCmd.Parsed() {
		cli.addressLists()
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
	fmt.Println("\tsend -from From -to To -amount Amount - 交易数据")
	fmt.Println("\tprintchain - 输出信息:")
	fmt.Println("\tgetbalance -address DATA -- 查询账户余额")
}
```

接下来就来测试一下嵌入地址和签名的转账机制

`main.go`

```go
package main

import "publicchain/cli"

func main() {
	cli := cli.CLI{}
	cli.Run()
}
```

删除之前的数据库

```
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go             
Usage:
        createwallet -- 创建钱包
        addresslists -- 输出所有钱包地址
        createblockchain -address DATA -- 创建创世区块
        send -from From -to To -amount Amount - 交易数据
        printchain - 输出信息:
        getbalance -address DATA -- 查询账户余额
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createwallet
创建钱包地址：19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createblockchain -address 19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3
数据库不存在,创建创世区块：
346: 0000781cada2d2bab1e8d359f07bfda37f40b7ea0665599f4d7a0368ca36e8b3
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go getbalance -address 19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3
查询余额： 19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3
19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3,一共有10个Token
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createwallet
创建钱包地址：18QXQHaa1NKffFqF3R2vj2fwuS7dvv76s2
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3"]' -to '["18QXQHaa1NKffFqF3R2vj2fwuS7dvv76s2"]' -amount '["5"]'
5058: 000080b1eab7768dc73899de0bf5a8556a8f7179e323e257166118775ad903e9
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3"]' -to '["18QXQHaa1NKffFqF3R2vj22"]' -amount '["5"]' 
钱包地址无效
Usage:
        createwallet -- 创建钱包
        addresslists -- 输出所有钱包地址
        createblockchain -address DATA -- 创建创世区块
        send -from From -to To -amount Amount - 交易数据
        printchain - 输出信息:
        getbalance -address DATA -- 查询账户余额
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go addresslists
打印所有的钱包地址
address: 1KL1V2xDtgj1ATTBfRNPxtd6SxYxPEQhN3
address: 1NfF1ThnrWKyQT8DndCEPDLp3kdQ1mjUqt
address: 19aLqRgFWz54Hhi1EAiVgfRTH28SaAL1e3
address: 18QXQHaa1NKffFqF3R2vj2fwuS7dvv76s2
```

嵌入地址的系统目前可以正常的完成所有的功能

# 十、更新优化

前面最开始构建了交易的基本模型，然后逐步实现了转账，集成了钱包地址。公链基本交易模块已然成型，还有些小的细节需要去实现和优化

## 1、UTXOset

之前为了实现转账引入了UTXO的概念，但是我们每次在转账查询可用余额时，都会去遍历一遍数据库上的区块。这样，会随着区块链的不断扩张，转账时查询的成本越来越高。

毕竟，查询时我们只需要关注未花费的TxOutput信息，而不需要关注区块上其他信息。那么，我们为什么不把未花费的TxOutput信息单独存储来查询呢？

其实，比特币正是这样做的。Bitcoin将链上所有区块存储在blocks数据库,将所有UTXO的集存储在chainstate数据库。

于是，我们引入UTXOSet集用来实现UTXO集的数据库存储

先在配置文件中设置utxotable的名字

`config.go`

```go
const UtxoTableName = "utxoTable" //UTXO的表名
```

在pbcc下创建utxo_set.go

创建UTXOSet结构体

`utxo_set.go`

```go
//UTXO结合结构体
type UTXOSet struct {
	BlockChain *BlockChain
}
```

重置utxo，该方法初始化UTXO集的存储表，如果bucket 存在就先移除，然后从区块链中获取所有的未花费输出，最终将输出保存到 bucket 中

```go
//重置UXTO_SET数据库表
func (utxoSet *UTXOSet) ResetUTXOSet() {
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.UtxoTableName))
		if b != nil {
			err := tx.DeleteBucket([]byte(conf.UtxoTableName))
			if err != nil {
				log.Panic("重置中，删除表失败")
			}

		}
		b, err := tx.CreateBucket([]byte(conf.UtxoTableName))
		if err != nil {
			log.Panic("重置中，创建新表失败")
		}
		if b != nil {
			txOutputMap := utxoSet.BlockChain.FindUnSpentOutputMap()
			//fmt.Println("未花费outputmap：",txOutputMap)
			for txIDStr, outputs := range txOutputMap {
				txID, _ := hex.DecodeString(txIDStr)
				b.Put(txID, outputs.Serilalize())
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

在blockchain.go中提供一个FindUnSpentOutputMap函数来帮助查询UTXO

我们存储UTXO的目的也是为了转账时能够更快地查询余额，为了实现转账，UTXO表除了存储对应的未花费的TXOutput集，还需要存储这些TXOutput来自于哪一笔交易

我们以交易的哈希为键，以该交易下TXOutput组成的数组为值来存储UTXO

由于一个交易下可能存在多个TXOutput，显然可以用数组表示。我们引入TXOutputs类来表示，因为要存储这些TXOutput需要实现序列化

该接口的作用就是存一笔交易的所有输出

`TXOutput.go`

```go
type TxOutputs struct {
	UTXOS []*UTXO
}

//序列化
func (outs *TxOutputs) Serilalize() []byte {
	//创建一个buffer
	var result bytes.Buffer
	//创建一个编码器
	encoder := gob.NewEncoder(&result)
	//编码--->打包
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TxOutputs {
	var txOutputs TxOutputs
	var reader = bytes.NewReader(txOutputsBytes)
	//创建一个解码器
	decoder := gob.NewDecoder(reader)
	//解包
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}
	return &txOutputs
}
```

知道UTXO表的存储格式后，我们就需要能够找到区块链上所有的未花费的TXOutput，并且能够以UTXO表存储的格式返回。这个方法是基于Blockchain的

`blockchain.go`

```go
//查询未花费的Output map[string] *TxOutputs
func (bc *BlockChain) FindUnSpentOutputMap() map[string]*TxOutputs {
	iterator := bc.Iterator()

	//存储已经花费：·[txID], txInput
	spentUTXOsMap := make(map[string][]*TXInput)
	//存储未花费
	unSpentOutputMaps := make(map[string]*TxOutputs)
	for {
		block := iterator.Next()
		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TxOutputs{[]*UTXO{}}
			tx := block.Txs[i]
			if !tx.IsCoinbaseTransaction() {
				for _, txInput := range tx.Vins {
					key := hex.EncodeToString(txInput.TxID)
					spentUTXOsMap[key] = append(spentUTXOsMap[key], txInput)
				}
			}
			txID := hex.EncodeToString(tx.TxID)
		work:
			for index, out := range tx.Vouts {
				txInputs := spentUTXOsMap[txID]
				if len(txInputs) > 0 {
					var isSpent bool
					for _, input := range txInputs {
						inputPubKeyHash := wallet.PubKeyHash(input.PublicKey)
						if bytes.Equal(inputPubKeyHash, out.PubKeyHash) {
							if input.Vout == index {
								isSpent = true
								continue work
							}
						}
					}
					if !isSpent {
						utxo := &UTXO{tx.TxID, index, out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}

				} else {
					utxo := &UTXO{tx.TxID, index, out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			//设置
			unSpentOutputMaps[txID] = txOutputs
		}

		//停止迭代
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unSpentOutputMaps
}
```

既然是UTXO存储表的初始化，那么一般情况下它只被执行一次。这种特性和创世区块相似，所以我们需要把他的实现放到创世区块的创建中

创建创世区块时候初始化UTXO表

```go
//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(address string) {
	if utils.DBExists() {
		fmt.Println("数据库已经存在")
		return
	}
	fmt.Println("数据库不存在,创建创世区块：")
	//先创建coinbase交易
	txCoinBase := NewCoinBaseTransaction(address)
	// 创世区块
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})
	//打开数据库
	db, err := bolt.Open(conf.DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(conf.BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误")
			}
			//存储最新区块的hash
			b.Put([]byte("l"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := &BlockChain{genesisBlock.Hash, db}
	utxoSet := &UTXOSet{bc}
	utxoSet.ResetUTXOSet()
}
```

## 2、转账优化

我们来回忆一下转账的几个重要步骤：

> 1.找到发起方所有的UTXO 
>
> 2.在1的UTXO集里找到符合该次转账条件的UTXO组合和该组合对应的代币总和
>
> 3.发起转账

之前上面的1，2都是在Blockchain里实现，并需要遍历整个区块数据库。在引入UTXOSet之后就需要基于UTXOSet实现。

> 1.Blockchain.UTXOs --> UTXOSet.FindUnPackageSpendableUTXOS

> 2.Blockchain.FindSpendableUTXOs --> UTXOSet.FindSpendableUTXOs

显然，2的方法基本相同。但是1有所差别，因为之前的逻辑中Blockchain.UTXOs需要找到链上所有UTXO和未打包的UTXO，而引入UTXOSet之后链上的UTXO都在UTXO表中，我们只需要找到未打包的交易产生的UTXO即可。

找到未打包交易的UTXO

`utxo_set.go`

```go
// 未打包的交易的UTXO
func (utxoSet *UTXOSet) FindUnPackageSpentableUTXOs(from string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO
	//存储已经花费
	spentTxOutput := make(map[string][]int)
	for i := len(txs) - 1; i >= 0; i-- {
		unUTXOs = caculate(txs[i], from, spentTxOutput, unUTXOs)
	}
	return unUTXOs
}
```

找到未花费交易里满足当次交易的UTXO组合，先找未打包的交易中的，钱不够再去数据库找

```go
//用于查询给定地址下的，要转账使用的可以使用的utxo
func (utxoSet *UTXOSet) FindSpendableUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	spentableUTXO := make(map[string][]int)
	var total int64 = 0
	//找出未打包的Transaction中未花费的
	unPackageUTXOs := utxoSet.FindUnPackageSpentableUTXOs(from, txs)
	for _, utxo := range unPackageUTXOs {
		total += utxo.Output.Value
		txIDStr := hex.EncodeToString(utxo.TxID)
		spentableUTXO[txIDStr] = append(spentableUTXO[txIDStr], utxo.Index)
		fmt.Println(amount, ",未打包，转账花费：", utxo.Output.Value)
		if total >= amount {
			return total, spentableUTXO
		}
	}
	//钱不够
	//找出已经存在数据库中的未花费的
	err := utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.UtxoTableName))
		if b != nil {
			c := b.Cursor()
		dbLoop:
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutpus := DeserializeTXOutputs(v)
				for _, utxo := range txOutpus.UTXOS {
					if utxo.Output.UnLockWithAddress(from) {
						total += utxo.Output.Value
						txIDStr := hex.EncodeToString(utxo.TxID)
						spentableUTXO[txIDStr] = append(spentableUTXO[txIDStr], utxo.Index)
						fmt.Println(amount, ",数据库，转账花费：", utxo.Output.Value)
						if total >= amount {
							break dbLoop
						}
					}

				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	if total < amount {
		fmt.Printf("%s,账户余额不足，不能转账。。", from)
		os.Exit(1)
	}
	return total, spentableUTXO
}
```

由于转账消耗了一定的UTXO，同时产生了一定的UTXO。所以转账之后需要对UTXO数据库表做更新以保持UTXO表存储的永远是最新的未花费的交易输出。

简单地梳理一下更新UTXO表的步骤:

> 1.找到最新添加到区块链上的区块
>
> 2.遍历区块交易，将所有交易输入集中到一个数组
>
> 3.遍历区块交易的交易输出，找到新增的未花费的TXOutput
>
> 4.在UTXO表中删除输入输入中已花费的TXOutput，并将未花费的TXOutput缓存
>
> 5.将3求出的和4缓存的TXOutput新增到UTXO表中

```go
//每次创建区块后(在这里就是每次交易以后)，更新未花费的表
func (utxoSet *UTXOSet) Update() {
	/*
		每当创建新区块后，都会花掉一些原来的utxo，产生新的utxo。
		删除已经花费的，增加新产生的未花费
		表中存储的数据结构：
		key：交易ID
		value：TxInputs
			TxInputs里是UTXO数组

	*/

	//获取最新的区块，由于该block的产生
	newBlock := utxoSet.BlockChain.Iterator().Next()

	inputs := []*TXInput{}
	outsMap := make(map[string]*TxOutputs)

	//获取已经花费的
	for _, tx := range newBlock.Txs {
		if tx.IsCoinbaseTransaction() {
			continue
		}
		// 把输入都拷贝下来
		for _, in := range tx.Vins {
			inputs = append(inputs, in)
		}
	}
	fmt.Println("inputs的长度:", len(inputs), inputs)
	//以上是找出新添加的区块中的所有的Input

	//以下是找到新添加的区块中的未花费了的Output
	for _, tx := range newBlock.Txs {
		utoxs := []*UTXO{}
	outLoop:
		for index, out := range tx.Vouts {
			isSpent := false
			for _, in := range inputs {
				if bytes.Equal(in.TxID, tx.TxID) && in.Vout == index && bytes.Equal(out.PubKeyHash, wallet.PubKeyHash(in.PublicKey)) {
					isSpent = true
					continue outLoop
				}
			}
			if !isSpent {
				utxo := &UTXO{tx.TxID, index, out}
				utoxs = append(utoxs, utxo)
				fmt.Println("新增UXTO", out.Value)
			}
		}
		// 如果这个交易中有UXTO那么组装一个map
		if len(utoxs) > 0 {
			txIDStr := hex.EncodeToString(tx.TxID)
			// 交易ID对应未花费的输出
			outsMap[txIDStr] = &TxOutputs{utoxs}
		}

	}
	fmt.Println("outsMap的长度:", len(outsMap), outsMap)

	//删除已经花费了的
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.UtxoTableName))
		if b != nil {
			//删除 ins中
			for i := 0; i < len(inputs); i++ {
				in := inputs[i]
				fmt.Println(i, "=========================")
				txOutputsBytes := b.Get(in.TxID)
				// 没有对应的交易的话
				if len(txOutputsBytes) == 0 {
					continue
				}
				txOutputs := DeserializeTXOutputs(txOutputsBytes)
				//根据IxID，如果该txOutputs中已经有output被新区块花掉了，那么将未花掉的添加到utxos里，并标记该txouputs要删除
				// 判断是否需要
				isNeedDelete := false
				utxos := []*UTXO{} //存储未花费
				for _, utxo := range txOutputs.UTXOS {
					if bytes.Equal(utxo.Output.PubKeyHash, wallet.PubKeyHash(in.PublicKey)) && in.Vout == utxo.Index {
						isNeedDelete = true
					} else {
						utxos = append(utxos, utxo)
					}
				}
				// 新的这笔交易的未花费长度
				fmt.Println(len(utxos))

				if isNeedDelete {
					b.Delete(in.TxID)
					if len(utxos) > 0 {
						txOutputs := &TxOutputs{utxos}
						b.Put(in.TxID, txOutputs.Serilalize())
						fmt.Println("删除时:map:", len(outsMap), outsMap)
					}

				}
			}
			//增加
			for keyID, outPuts := range outsMap {
				keyHashBytes, _ := hex.DecodeString(keyID)
				b.Put(keyHashBytes, outPuts.Serilalize())
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

到现在转账就可以借用UTXO_SET来优化了

## 3、铸币奖励

我们知道，在比特币中每当矿工成功挖到一个区块，就会得到一笔奖励。这笔奖励包含在铸币交易中，铸币交易是一个区块中的第一笔交易

目前的项目还没有引入多节点竞争挖矿，暂且认为每一个区块是转账的第一个发起人挖到的。如此，Coinbase奖励就应该添加到MineNewBlock方法里

```go
//挖掘新的区块 有交易的时候就会调用
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction
	//奖励
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)

	utxoSet := &UTXOSet{bc}

	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		tx := NewSimpleTransaction(from[i], to[i], amountInt, utxoSet, txs)
		txs = append(txs, tx)
	}

	var block *Block    //数据库中的最后一个block
	var newBlock *Block //要创建的新的block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) //数据库中的最后一个block
		}
		return nil
	})

	//在建立新区块钱，对txs进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx, _txs) != true {
			log.Panic("签名验证失败。。")
		}
		_txs = append(_txs, tx)
	}

	newBlock = NewBlock(txs, block.Hash, block.Height+1)

	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
}
```

这个时候我们的一般交易的还没有使用优化过的utxo_set，下面我们来改动一下

`transaction.go`

```go
func NewSimpleTransaction(from, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction) *Transaction {
	var txInputs []*TXInput
	var txOutputs []*TXOuput
	balance, spendableUTXO := utxoSet.FindSpendableUTXOs(from, amount, txs)

	//获取钱包
	wallets := wallet.NewWallets()
	wallet := wallets.WalletsMap[from]

	for txID, indexArray := range spendableUTXO {
		txIDBytes, _ := hex.DecodeString(txID)
		for _, index := range indexArray {
			txInput := &TXInput{txIDBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput1 := NewTXOuput(amount, to)
	txOutputs = append(txOutputs, txOutput1)

	//找零
	txOutput2 := NewTXOuput(balance-amount, from)
	txOutputs = append(txOutputs, txOutput2)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.SetTxID()

	//进行签名
	utxoSet.BlockChain.SignTransaction(tx, wallet.PrivateKey, txs)
	return tx
}
```

## 4、获取余额

现在我们查询余额就不需要去遍历整个区块数据库了，只需要遍历存储UTXO的表即可，查询逻辑还是和之前一样，先要从表中查到对应地址的所有UTXO，然后累加他们的值

`utxo_set.go`

```go
// 获取地址余额
func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	utxos := utxoSet.FindUnspentOutputsForAddress(address)
	var amount int64
	for _, utxo := range utxos {
		amount += utxo.Output.Value
		fmt.Println(address, "余额：", utxo.Output.Value)
	}
	//fmt.Printf("%s账户，有%d个Token\n",address,amount)
	return amount
}
// 找到对应地址的所有UTXO
func (utxoSet *UTXOSet) FindUnspentOutputsForAddress(address string) []*UTXO {
	var utxos []*UTXO
	//查询数据，遍历所有的未花费
	err := utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.UtxoTableName))
		if b != nil {
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				//fmt.Printf("key=%s,value=%v\n", k, v)
				txOutputs := DeserializeTXOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					if utxo.Output.UnLockWithAddress(address) {
						utxos = append(utxos, utxo)
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return utxos
}
```

## 5、设置命令行参数

首先要对转账进行修改

`cli_send.go`

```go
//转账
func (cli *CLI) send(from, to, amount []string) {
	blockchain := pbcc.GetBlockchainObject()
	if blockchain == nil {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}
	blockchain.MineNewBlock(from, to, amount)
	defer blockchain.DB.Close()

	utxoSet := &pbcc.UTXOSet{BlockChain: blockchain}
	//转账成功以后，需要更新
	utxoSet.Update()
}
```

获取地址余额的方法也要修改

`cli_getbalance.go`

```go
//查询余额
func (cli *CLI) getBalance(address string) {
	fmt.Println("查询余额：", address)
	bc := pbcc.GetBlockchainObject()

	if bc == nil {
		fmt.Println("数据库不存在，无法查询。。")
		os.Exit(1)
	}
	defer bc.DB.Close()
	utxoSet := &pbcc.UTXOSet{BlockChain: bc}
	balance := utxoSet.GetBalance(address)
	fmt.Printf("%s,一共有%d个Token\n", address, balance)
}
```

同时我们也可以设置一个输出所有的UTXO来验证一下

`cli_utxoset.go`

```go
// 输出UXTO
func (cli *CLI) TestMethod() {
	blockchain := pbcc.GetBlockchainObject()
	defer blockchain.DB.Close()
	unSpentOutputMap := blockchain.FindUnSpentOutputMap()
	fmt.Println(unSpentOutputMap)
	for key, value := range unSpentOutputMap {
		fmt.Println(key)
		for _, utxo := range value.UTXOS {
			fmt.Println("金额：", utxo.Output.Value)
			fmt.Printf("地址：%v\n", utxo.Output.PubKeyHash)
			fmt.Println("---------------------")
		}
	}
	utxoSet := &pbcc.UTXOSet{BlockChain: blockchain}
	utxoSet.ResetUTXOSet()
}
```

更新cli.go

`cli.go`

```go
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

	//创建flagset标签对象
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	addressListsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)

	//设置标签后的参数
	flagFromData := sendBlockCmd.String("from", "", "转帐源地址")
	flagToData := sendBlockCmd.String("to", "", "转帐目标地址")
	flagAmountData := sendBlockCmd.String("amount", "", "转帐金额")
	flagCreateBlockChainData := createBlockChainCmd.String("address", "", "创世区块交易地址")
	flagGetBalanceData := getBalanceCmd.String("address", "", "要查询的某个账户的余额")

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

		cli.send(from, to, amount)
	}
	if printChainCmd.Parsed() {
		cli.printChains()
	}

	if createBlockChainCmd.Parsed() {
		if !wallet.IsValidForAddress([]byte(*flagCreateBlockChainData)) {
			fmt.Println("创建地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockChainData)
	}

	if getBalanceCmd.Parsed() {
		if !wallet.IsValidForAddress([]byte(*flagGetBalanceData)) {
			fmt.Println("查询地址无效")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceData)

	}

	if createWalletCmd.Parsed() {
		//创建钱包
		cli.createWallet()
	}
	//获取所有的钱包地址
	if addressListsCmd.Parsed() {
		cli.addressLists()
	}
	if testCmd.Parsed() {
		cli.TestMethod()
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
	fmt.Println("\tsend -from From -to To -amount Amount - 交易数据")
	fmt.Println("\tprintchain - 输出信息:")
	fmt.Println("\tgetbalance -address DATA -- 查询账户余额")
	fmt.Println("\ttest -- 测试")
}
```

下面继续来测试一下优化过的的BTC系统吧

main.go不变

同样还是删除之前的数据库和钱包集

```go
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go 
Usage:
        createwallet -- 创建钱包
        addresslists -- 输出所有钱包地址
        createblockchain -address DATA -- 创建创世区块
        send -from From -to To -amount Amount - 交易数据
        printchain - 输出信息:
        getbalance -address DATA -- 查询账户余额
        test -- 测试
exit status 1
(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createwallet
文件不存在
创建钱包地址：1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createblockchain -address 1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu
数据库不存在,创建创世区块：
202840: 000073404ca4f64f73316ab16fad9faffb4e0b506e883447745d1d663f5f4622

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go getbalance -address 1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu
查询余额： 1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu
1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu 余额： 10
1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu,一共有10个Token

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go createwallet
创建钱包地址：1AQ3Ndj6yP6VZ52q5mkVAwF45DkSrfQb8t

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go send -from '["1DRy1H3UKgHA1dFDmfsTEa28Sy8g8fjHLu"]' -to '["1AQ3Ndj6yP6VZ52q5mkVAwF45DkSrfQb8t"]' -amount '["5"]'
5 ,未打包，转账花费： 10
11550: 0000d96097d90635ad1b40c6a8280635c20587ea894064af1fc3232522c112ec
inputs的长度: 1 [0xc00006a2d0]
outsMaps, 5
outsMaps, 5
outsMap的长度: 1 map[4b290e66ec6f920a88d38ce750dcdffd7949c02c31e4f7bd453d1f1c5fd02746:0xc00018fb30]
0 =========================

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go test
map[098d9ed2a17e54a72457411e9d61efd442d82a06035c4a000864ef9fabcbe804:0xc0000acb10 4b290e66ec6f920a88d38ce750dcdffd7949c02c31e4f7bd453d1f1c5fd02746:0xc0000acaf8 f77ad188275786391d0a856e17b4fc6aa02ad72f41b5c0e95b151eab52b6bb71:0xc0000acf30]
4b290e66ec6f920a88d38ce750dcdffd7949c02c31e4f7bd453d1f1c5fd02746
金额： 5
地址：[103 20 174 23 99 26 82 53 183 229 119 88 18 64 178 163 171 119 156 185]
---------------------
金额： 5
地址：[136 90 62 47 173 206 217 154 25 141 105 215 22 177 58 76 162 34 25 137]
---------------------
098d9ed2a17e54a72457411e9d61efd442d82a06035c4a000864ef9fabcbe804
f77ad188275786391d0a856e17b4fc6aa02ad72f41b5c0e95b151eab52b6bb71
金额： 10
地址：[136 90 62 47 173 206 217 154 25 141 105 215 22 177 58 76 162 34 25 137]
---------------------

(base) yunphant@yunphantdeMacBook-Pro publicchain % go run main.go printchain
第2个区块的信息:
        高度:1
        上一个区块的hash:000073404ca4f64f73316ab16fad9faffb4e0b506e883447745d1d663f5f4622
        当前的hash:0000d96097d90635ad1b40c6a8280635c20587ea894064af1fc3232522c112ec
        交易:
                交易ID:098d9ed2a17e54a72457411e9d61efd442d82a06035c4a000864ef9fabcbe804
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[136 90 62 47 173 206 217 154 25 141 105 215 22 177 58 76 162 34 25 137]
                交易ID:4b290e66ec6f920a88d38ce750dcdffd7949c02c31e4f7bd453d1f1c5fd02746
                Vins:
                        TxID:098d9ed2a17e54a72457411e9d61efd442d82a06035c4a000864ef9fabcbe804
                        Vout:0
                        PublicKey:[191 96 147 35 114 116 88 165 31 247 180 159 93 237 219 103 146 68 5 21 111 121 181 180 12 38 125 41 164 255 239 35 107 221 57 38 35 79 81 27 181 25 45 110 212 19 80 210 89 89 177 45 90 113 114 120 15 175 237 160 0 102 243 30]
                Vouts:
                        value:5
                        PubKeyHash:[103 20 174 23 99 26 82 53 183 229 119 88 18 64 178 163 171 119 156 185]
                        value:5
                        PubKeyHash:[136 90 62 47 173 206 217 154 25 141 105 215 22 177 58 76 162 34 25 137]
        时间:2022-07-11 17:29:42
        次数:11550
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:000073404ca4f64f73316ab16fad9faffb4e0b506e883447745d1d663f5f4622
        交易:
                交易ID:f77ad188275786391d0a856e17b4fc6aa02ad72f41b5c0e95b151eab52b6bb71
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[136 90 62 47 173 206 217 154 25 141 105 215 22 177 58 76 162 34 25 137]
        时间:2022-07-11 17:27:02
        次数:202840
```

完全正确，到目前为止，公链的交易算是基本讲完了

# 十一、MerkleTree

我们完成了该教程的上半部分，在进入下一部分前，作为过渡章节，我期望再为我们的goblockchain添加一些常见的功能，就比如本文即将介绍的区块中的Merkle Tree（梅克尔树，后面简称MT）和区块链系统的SPV（快速交易验证）功能。对于接触了区块链一段时间的读者应该都有了解过MT，尽管MT的数据结构非常容易理解，但是我认为MT对于区块链的意义以及其涉及的SPV功能是认知整个区块链技术的关键步骤之一,通过对MT与SPV的深入理解与思考，我们也许能够窥探区块链在未来的应用场景以及更好地判断区块链技术当前面临的技术瓶颈。同时我相信大多数读者对于MT在区块链中的具体实现以及如何通过MT树实现SPV的过程是存在理解偏差的，本文将通过代码的形式讲述MT与SPV的实现，希望读者能够重点关注一些唯独在具体实现才会遇见的问题，这些隐藏在代码里的问题的解决可以帮助读者宏观把控区块链系统的设计理念，这也是本教程的初衷。

那么MerkleTree对我们构造公链有什么用呢？

我们知道完整的比特币数据库已达到一百多Gb的存储，对于每一个节点必须保存一个区块链的完整副本。这对于大多数使用比特币的人显然不合适，于是中本聪提出了简单支付验证SPV(Simplified Payment Verification).

简单地说，SPV是一个轻量级的比特币节点，它并不需要下载区块链的所有数据内容。为了实现SPV，就需要有一个方式来检查某区块是否包含某一笔交易，这就是MerkleTree能帮我们解决的问题。

一个区块的结构里只有一个哈希值，但是这个哈希值包含了所有交易的哈希值。我们将区块内所有交易哈希的值两两进行哈希得到一个新的哈希值，然后再把得到的新的哈希值两两哈希...不断进行这个过程直到最后只存在一个哈希值。这样的结构是不是很像一颗二叉树，我们将这样的二叉树就叫做MerkleTree。

比特币的MerkleTree结构图：

值，然后再把得到的新的哈希值两两哈希...不断进行这个过程直到最后只存在一个哈希值。这样的结构是不是很像一颗二叉树，我们将这样的二叉树就叫做MerkleTree。

这样，我们只需要一个根哈希就可以验证一笔交易是否存在于一个区块中了，因为这个根哈希可以遍历到所有交易哈希。相当于一个Merkle 树根哈希和一个 Merkle 路径

## 1、结构

MerkleTree只包含一个根节点，每一个默克尔树节点包含数据和左右指针。每个节点都可以连接到下个节点，并依次连接到更远的节点直到叶子节点

`merkletree.go`

```go
//默克尔树
type MerkleTree struct {
	RootNode *MerkleNode //根节点
}

//默克尔树节点
type MerkleNode struct {
	Left  *MerkleNode //左节点
	Right *MerkleNode //右节点
	Data  []byte      //节点数据
}
```

新建节点时首先要从叶子节点开始创建，叶子节点只有数据，对交易哈希进行哈希得到叶子节点的数值；当创建非叶子节点时，将左右子节点的数据拼接进行哈希得到新节点的数据值

`merkletree.go`

```go
//创建节点
func NewMerkleNode(leftNode, rightNode *MerkleNode, txHash []byte) *MerkleNode {
	//创建当前的节点
	mNode := &MerkleNode{}
	//赋值
	if leftNode == nil && rightNode == nil {
		//mNode就是个叶子节点
		hash := sha256.Sum256(txHash)
		mNode.Data = hash[:]
	} else {
		//mNOde是非叶子节点
		prevHash := append(leftNode.Data, rightNode.Data...)
		hash := sha256.Sum256(prevHash)
		mNode.Data = hash[:]
	}
	mNode.LeftNode = leftNode
	mNode.RightNode = rightNode
	return mNode
}
```

当用叶子节点去生成一颗默克尔树时，必须保证叶子节点的数量为偶数，如果不是需要复制一份最后的交易哈西值到最后拼凑成偶数个交易哈希

叶子节点两两哈希形成新的节点，新节点继续两两哈希直到最后只有一个节点

`merkletree.go`

```go
//创建merkletree
func NewMerkleTree(txHashData [][]byte) *MerkleTree {
	//创建一个数组，用于存储node节点
	var nodes []*MerkleNode
	//判断交易量的奇偶性
	if len(txHashData)%2 != 0 {
		//奇数，复制最后一个
		txHashData = append(txHashData, txHashData[len(txHashData)-1])
	}
	//创建一排的叶子节点
	for _, datum := range txHashData {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, node)
	}
	// 计算构建一棵树需要循环的次数
	count := GetCircleCount(len(nodes))
	for i := 0; i < count; i++ {
		var newLevel []*MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			newLevel = append(newLevel, node)

		}

		//判断newLevel的长度的奇偶性
		if len(newLevel)%2 != 0 {
			newLevel = append(newLevel, newLevel[len(newLevel)-1])
		}
		nodes = newLevel

	}
	mTree := &MerkleTree{nodes[0]}
	return mTree
}

//获取产生Merkle树根需要循环的次数
func GetCircleCount(len int) int {
	count := 0
	for {
		if int(math.Pow(2, float64(count))) >= len {
			return count
		}
		count++
	}
}
```

## 2、集成到区块中

到现在我们已经理解merkletree了，其实在BTC中，区块的信息也是通过merkletree去构建的，但是我们这里就进行对区块内交易的merkletree的构建

`block.go`

```go
//将Txs转为[]byte
func (block *Block) HashTransactions() []byte {
	var txs [][]byte
	for _, tx := range block.Txs {
		txs = append(txs, tx.Serialize())
	}
	mTree := NewMerkleTree(txs)
	return mTree.RootNode.Data
}
```

之前值直接把所有的交易都整合在一起，然后取hash获得该区块所有的交易的hash

现在我们获取所有交易，然后使用所有的交易构建一个merkletree，这样就实现了区块内merkletree的构建

本章节测试的想法还没想好，后面再补充

# 十二、P2P网络模拟

接下来实现BTC的P2P网络模拟，本节的知识我们用一台机器开设三个端口模拟三台机器，（也可以自己用三台机器去实现）

前面实现了公链的基本结构，交易，钱包地址，数据持久化，交易等功能。但显然这些功能都是基于单节点的，我们都知道比特币网络是一个多节点共存的P2P网络，而且网络中的节点也是有分类的的，主要是分成一下三类

矿工节点：具备挖矿功能的节点。这些节点一般运行在特殊的硬件设备以完成复杂的工作量证明运算。有些矿工节点同时也是全节点

钱包节点：常见的很多比特币客户端属于钱包节点，它不需要拷贝完整的区块链。一般的钱包节点都是SPV节点，SPV节点借助之前讲的MerkleTree原理使得不需要下载所有区块就能验证交易成为可能，后面讲到钱包开发再深入理解

全节点：具有完整的，最新的区块链拷贝。可以独立自主地校验所有交易



由于P2P网络的复杂性，为了便于理解区块链网络同步的原理，我们在接下来的实验中就搭建三类节点（和BTC中的节点略有不同）

中心节点(全节点)：其他节点会连接到这个节点来更新区块数据

钱包节点：用于钱包之间实现交易，但这里它依旧存储一个区块链的完整副本，注意是完整的

矿工节点：矿工节点会在内存池中存储交易并在适当时机将交易打包挖出一个新区块 但这里它依旧存储一个区块链的完整副本

我们在这个简化基础上去实现区块链的网络同步



创建server包，在server包下创建server.go\server_model.go\server_send.go\server_handle.go\server_var.go

**本节的代码改动量比之前都要多，大家仔细操作**

## 1、新增接口

继续我们要先在原来的基础上添加几个新的接口

`utils.go`

```go
//消息类型转字节数组
func CommandToBytes(command string) []byte {
	// 规定了消息类型的字节长度
	var bytes [conf.COMMANDLENGTH]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

//字节数组转消息类型
func BytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

// 将结构体序列化成字节数组
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
```

`blockchain.go`

```go
//P2P新增接口
//获取最新区块的高度
func (bc *BlockChain) GetBestHeight() int64 {
	block := bc.Iterator().Next()
	return block.Height
}

//获取所有区块的hash
func (bc *BlockChain) GetBlockHashes() [][]byte {
	blockIterator := bc.Iterator()
	var blockHashs [][]byte
	for {
		block := blockIterator.Next()
		blockHashs = append(blockHashs, block.Hash)
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return blockHashs
}

//根据hash获取区块
func (bc *BlockChain) GetBlock(blockHash []byte) ([]byte, error) {
	var blockBytes []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			blockBytes = b.Get(blockHash)
		}
		return nil
	})
	return blockBytes, err
}

//添加区块到数据库
func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			blockExist := b.Get(block.Hash)
			if blockExist != nil {
				// 如果存在，不需要做任何过多的处理
				return nil
			}
			err := b.Put(block.Hash, block.Serilalize())
			if err != nil {
				log.Panic(err)
			}
			// 最新的区块链的Hash
			blockHash := b.Get([]byte("l"))
			blockBytes := b.Get(blockHash)
			blockInDB := DeserializeBlock(blockBytes)
			if blockInDB.Height < block.Height {
				b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

到现在会有一个报错，暂时不用管，一会添加完配置就会解决了

## 2、消息分类

要想实现数据的同步，必须有两个节点间的通讯。那么他们通讯的内容和格式是什么样的呢？

区块链同步时两个节点的通讯信息并不是单一的，不同的情况和不同的阶段通讯的格式与处理方式是不同的。这里分析主要用的几个数据结构。

为了区分节点发送的信息，我们需要定义几个消息类型来区别他们

`config.go`

```go
const PROTOCOL = "tcp"   // 采用TCP
const COMMANDLENGTH = 12 // 发送消息的前12个字节指定了命令名(version)
const NODE_VERSION = 1   // 节点的区块链版本

// 命令
const COMMAND_VERSION = "version"     //该消息的目的就是比较谁的链长
const COMMAND_ADDR = "addr"           //消息没有实现具体的业务
const COMMAND_BLOCK = "block"         //该消息是发送一个区块
const COMMAND_INV = "inv"             //该消息是把自己的区块信息和
const COMMAND_GETBLOCKS = "getblocks" //该消息是请求获取对方的所有区块的hash
const COMMAND_GETDATA = "getdata"     //该消息是请求获取区块或者Tx的信息
const COMMAND_TX = "tx"               //该消息是发送交易

// 类型 用于区分Inv消息发送的是区块还是交易
const BLOCK_TYPE = "block"
const TX_TYPE = "tx"
```

同时我们也要添加一些节点同步的全局变量

`server_var.go`

```go
//存储节点全局变量
var KnowNodes = []string{"localhost:8000"}            //localhost:3000 主节点的地址
var NodeAddress string                                //全局变量，节点地址
var TransactionArray [][]byte                         // 存储hash值
var MinerAddress string                               //旷工地址
var MemoryTxPool = make(map[string]*pbcc.Transaction) //交易池存储交易
```

## 3、version消息

Version消息是发起区块同步第一个发送的消息类型，其内容主要有区块链版本，区块链最大高度，来自的节点地址。它主要用于比较两个节点间谁是最长链

`server_model.go`

```go
//version消息结构体
type Version struct {
	Version    int64  // 版本
	BestHeight int64  // 当前节点区块的高度
	AddrFrom   string //当前节点的地址
}
```

组装发送Version信息

`server_send.go`

```go
//组装版本消息数据
func SendVersion(toAddress string, blc *pbcc.BlockChain) {
	// 获取获取区块高度
	bestHeight := blc.GetBestHeight()
	// 组装版本数据
	payload := utils.GobEncode(Version{conf.NODE_VERSION, bestHeight, NodeAddress})
	// 把命令和数据组成请求
	request := append(utils.CommandToBytes(conf.COMMAND_VERSION), payload...)
	// 数据发送
	SendData(toAddress, request)
}

// 像其他节点发送数据
func SendData(to string, data []byte) {
	// 获取链接对象
	conn, err := net.Dial(conf.PROTOCOL, to)
	if err != nil {
		panic("error")
	}
	defer conn.Close()
	// 附带要发送的数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}
```

当一个节点收到Version信息，会比较自己的最大区块高度和请求者的最大区块高度。如果自身高度大于请求节点会向请求节点回复一个版本信息告诉请求节点自己的相关信息；否则直接向请求节点发送一个GetBlocks信息

`server_handle.go`

```go
// 处理版本消息
func handleVersion(request []byte, bc *pbcc.BlockChain) {

	var buff bytes.Buffer
	var payload Version
	// 从请求中截出数据
	dataBytes := request[conf.COMMANDLENGTH:]

	// 反序列化 解析请求数据中的version消息到payload
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// 获取本节点存的链的区块高度
	bestHeight := bc.GetBestHeight()
	// 节点请求发来消息的区块高度
	foreignerBestHeight := payload.BestHeight
	// 如果本节点的区块高度大于发来消息节点的区块高度
	if bestHeight > foreignerBestHeight {
		//把本节点的区块链高度信息发给对方
		SendVersion(payload.AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		// 去向对方节点获取区块
		SendGetBlocks(payload.AddrFrom)
	}
	// 如果该节点之前没来同步过，那么加入已知节点的列表
	if !nodeIsKnown(payload.AddrFrom) {
		KnowNodes = append(KnowNodes, payload.AddrFrom)
	}
}
```

到目前会有一些报错，不要管，继续往下

## 4、GetBlocks

如果本节点收到对方节点发来的Version消息，那么当自己节点的区块高度小于对方的时候，那么那要去向对方请求区块消息

一般收到GetBlocks消息的节点为较新区块链

`server_model.go`

```go
//请求区块信息 意为 “给我看一下你有什么区块”（在比特币中，这会更加复杂）
type GetBlocks struct {
	AddrFrom string
}
```

组装发送GetBlock消息

`server_send.go`

```go
//组装获取区块消息并发送
func SendGetBlocks(toAddress string) {
	// 指定从全节点获取
	payload := utils.GobEncode(GetBlocks{NodeAddress})
	// 拼接命令和消息
	request := append(utils.CommandToBytes(conf.COMMAND_GETBLOCKS), payload...)
	fmt.Printf("像节点地址为:%s的节点发送了GetBlock消息", toAddress)
	SendData(toAddress, request)
}
```

同时提供一个判断节点是否为已知节点的接口

`server.go`

```go
// 判断节点是否为已知节点
func nodeIsKnown(addr string) bool {
	for _, node := range KnowNodes {
		if node == addr {
			return true
		}
	}
	return false
}
```

当一个节点收到一个GetBlocks消息，会将自身区块链所有区块哈希算出并组装在Inv消息中发送给请求节点。一般收到GetBlocks消息的节点为较新区块链

`server_handle.go`

```go
// 处理GetBlock消息
func handleGetblocks(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//获取所有区块的hash
	blocks := bc.GetBlockHashes()
	//向请求地址发送Inv消息
	SendInv(payload.AddrFrom, conf.BLOCK_TYPE, blocks)
}
```

## 5、Inv消息

Inv消息用于收到GetBlocks消息的节点向其他节点展示自己拥有的区块或交易信息。其主要结构包括自己的节点地址，展示信息的类型，是区块还是交易，当用于节点请求区块同步时是区块信息；当用于节点向矿工节点转发交易时是交易信息

`server_inv.go`

```go
// Inv消息结构体 像别人展示自己的区块或者交易的信息
type Inv struct {
	AddrFrom string   //自己的地址
	Type     string   //类型 block tx
	Items    [][]byte //hash二维数组
}
```

组装Inv消息并且发送

`servr_send.go`

```go
// 组装Inv消息并发送
func SendInv(toAddress string, kind string, hashes [][]byte) {
	// 从全节点获取
	payload := utils.GobEncode(Inv{NodeAddress, kind, hashes})
	// 拼接命令和数据
	request := append(utils.CommandToBytes(conf.COMMAND_INV), payload...)
	SendData(toAddress, request)
}
```

当一个节点收到Inv消息后，会对Inv消息的类型做判断分别采取处理。 如果是Block类型，它会取出最新的区块哈希并组装到一个GetData消息返回给来源节点，这个消息才是真正向来源节点请求新区块的消息

由于这里将源节点(比当前节点拥有更新区块链的节点)所有区块的哈希都知道了，所以需要每处理一次Inv消息后将剩余的区块哈希缓存到unslovedHashes数组，当unslovedHashes长度为零表示处理完毕

这里可能有人会有疑问，我们更新的应该是源节点拥有的新区块(自身节点没有)，这里为啥请求的是全部呢？这里的逻辑是这样的，请求的时候是请求的全部，后面在真正更新自身数据库的时候判断是否为新区块并保存到数据库。其实，我们都知道两个节点的区块最大高度，这里也可以完全请求源节点的所有新区块哈希。为了简单，这里先暂且这样处理

如果收到的Inv是交易类型，取出交易哈希，如果该交易不存在于交易缓冲池，添加到交易缓冲池。这里的交易类型Inv一般用于有矿工节点参与的通讯。因为在网络中，只有矿工节点才需要去处理交易

`server_handle.go`

```go
// 处理Inv消息
func handleInv(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload Inv
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// 如果Inv消息的数据是Block类型
	if payload.Type == conf.BLOCK_TYPE {
		// 记录最新的区块hash
		blockHash := payload.Items[0]
		// 发送GetDate消息
		sendGetData(payload.AddrFrom, conf.BLOCK_TYPE, blockHash)
		// 如果携带的区块或者交易数量大于1
		if len(payload.Items) >= 1 {
			//存下其他剩余区块的hash
			TransactionArray = payload.Items[1:]
		}
	}
	// 如果Inv消息的数据是Tx类型
	if payload.Type == conf.TX_TYPE {
		// 获取最后一笔交易
		txHash := payload.Items[0]
		// 如果缓冲交易池里面没有这个交易，则像节点发送GetData数据
		if MemoryTxPool[hex.EncodeToString(txHash)] == nil {
			SendGetData(payload.AddrFrom, conf.TX_TYPE, txHash)
		}
	}
}
```

## 6、GetData消息

GetData消息是用于真正请求一个区块或交易的消息类型，其主要结构为

`server_model.go`

```go
// GetData消息结构体  用于某个块或交易的请求，它可以仅包含一个块或交易的ID。
type GetData struct {
	AddrFrom string
	Type     string
	Hash     []byte //获取的是hash
}
```

组装并发送GetData消息

`server_send.go`

```go
// 组装GetData消息并发送
func SendGetData(toAddress string, kind string, blockHash []byte) {
	// 向全节点获取
	payload := utils.GobEncode(GetData{NodeAddress, kind, blockHash})
	request := append(utils.CommandToBytes(conf.COMMAND_GETDATA), payload...)
	SendData(toAddress, request)
}
```

当一个节点收到GetData消息，如果是请求区块，节点会根据区块哈希取出对应的区块封装到BlockData消息中发送给请求节点；如果是请求交易，同理会根据交易哈希取出对应交易封装到TxData消息中发送给请求节点

`server_handle.go`

```go
// 处理GetData消息
func handleGetData(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload GetData
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == conf.BLOCK_TYPE {
		// 获取区块消息
		block, err := bc.GetBlock([]byte(payload.Hash))
		if err != nil {
			return
		}
		SendBlock(payload.AddrFrom, block)
	}
	if payload.Type == conf.TX_TYPE {
		tx := MemoryTxPool[hex.EncodeToString(payload.Hash)]
		SendTx(payload.AddrFrom, tx)
	}
```

## 7、BlockData消息

BlockData消息用于一个节点向其他节点发送一个区块，到这里才真正完成区块的发送

`server_model.go`

```go
// BlockData消息结构体 给发送GetData请求回复区块
type BlockData struct {
	AddrFrom string
	Block    []byte
}
```

组装并发送BlockData消息

`server_send.go`

```go
// 组装BlockData消息并发送
func SendBlock(toAddress string, block []byte) {
	payload := utils.GobEncode(BlockData{NodeAddress, block})
	request := append(utils.CommandToBytes(conf.COMMAND_BLOCK), payload...)
	SendData(toAddress, request)
}
```

当一个节点收到一个Block信息，它会首先判断是否拥有该Block，如果数据库没有就将其添加到数据库中(AddBlock方法)。然后会判断unslovedHashes(之前缓存所有主节点未发送的区块哈希数组)数组的长度，如果数组长度不为零表示还有未发送处理的区块，节点继续发送GetData消息去请求下一个区块。否则，区块同步完成，重置UTXO数据库

`server_handle.go`

```go
// 处理发送区块消息
func handleBlock(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload BlockData
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockBytes := payload.Block
	// 解析获取区块
	block := pbcc.DeserializeBlock(blockBytes)
	fmt.Println("Recevied a new block!")
	// 新的区块加入链上
	bc.AddBlock(block)
	fmt.Printf("Added block %x\n", block.Hash)
	// 如果还有区块
	if len(TransactionArray) > 0 {
		blockHash := TransactionArray[0]
		// 再去请求
		SendGetData(payload.AddrFrom, "block", blockHash)
		// 更新未打包进区块链的区块池
		TransactionArray = TransactionArray[1:]
	} else {
		fmt.Println("已经没有要处理的区块了")
	}
}
```

## 8、TxData消息

TxData消息用于真正地发送一笔交易。当对方节点发送的GetData消息为Tx类型，相应地会回复TxData消息

`server_model.go`

```go
// Tx消息结构体 给发送GetData请求回复交易
type Tx struct {
	AddrFrom string
	Tx       *pbcc.Transaction
}
```

组装并发送TxData消息

`server_send.go`

```go
// 组装BTXData消息并发送
func SendTx(toAddress string, tx *pbcc.Transaction) {
	payload := utils.GobEncode(Tx{NodeAddress, tx})
	request := append(utils.CommandToBytes(conf.COMMAND_TX), payload...)
	SendData(toAddress, request)
}
```

当一个节点收到TxData消息，这个节点一般为矿工节点，如果不是他会以Inv消息格式继续转发该交易信息到矿工节点。矿工节点收到交易，当交易池满足一定数量时开始打包挖矿。

当生成新的区块并打包到区块链上时，矿工节点需要以BlockData消息向其他节点转发该新区块

`server_handle.go`

```go
// 处理发送交易消息
func handleTx(request []byte, bc *pbcc.BlockChain) {
	var buff bytes.Buffer
	var payload Tx
	dataBytes := request[conf.COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	tx := payload.Tx
	// 交易存到交易缓冲池子
	MemoryTxPool[hex.EncodeToString(tx.TxID)] = tx
	// 说明主节点自己
	if NodeAddress == KnowNodes[0] {
		// 给矿工节点发送交易hash
		for _, nodeAddr := range KnowNodes {
			if nodeAddr != NodeAddress && nodeAddr != payload.AddrFrom {
				SendInv(nodeAddr, conf.TX_TYPE, [][]byte{tx.TxID})
			}
		}
	}
	// 矿工进行挖矿验证
	if len(MemoryTxPool) >= 1 && len(MinerAddress) > 0 {
	MineTransactions:
		utxoSet := &pbcc.UTXOSet{bc}
		txs := []*pbcc.Transaction{tx}
		//奖励
		coinbaseTx := pbcc.NewCoinBaseTransaction(MinerAddress)
		txs = append(txs, coinbaseTx)
		_txs := []*pbcc.Transaction{}
		for _, tx := range txs {
			// 数字签名失败
			if bc.VerifyTransaction(tx, _txs) != true {
				log.Panic("ERROR: Invalid transaction")
			}
			_txs = append(_txs, tx)
		}
		//通过相关算法建立Transaction数组
		var block *pbcc.Block
		bc.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
			if b != nil {
				hash := b.Get([]byte("l"))
				blockBytes := b.Get(hash)
				block = pbcc.DeserializeBlock(blockBytes)
			}
			return nil
		})
		//建立新的区块
		block = pbcc.NewBlock(txs, block.Hash, block.Height+1)
		//将新区块存储到数据库
		bc.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
			if b != nil {
				b.Put(block.Hash, block.Serilalize())
				b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
			return nil
		})
		utxoSet.Update()
		SendBlock(KnowNodes[0], block.Serilalize())
		for _, tx := range txs {
			txID := hex.EncodeToString(tx.TxID)
			delete(MemoryTxPool, txID)
		}
		for _, node := range KnowNodes {
			if node != NodeAddress {
				SendInv(node, "block", [][]byte{block.Hash})
			}
		}
		if len(MemoryTxPool) > 0 {
			goto MineTransactions
		}
	}
}
```

到这里网络通信消息的类型已经搭建完毕，到目前为止可能感受不够直观，下面搭建服务器后面就能更好理解

## 9、server服务器

由于我们是在本地模拟网络环境，所以采用不同的端口号来模拟节点IP地址。eg：localhost:8000代表一个节点，eg：localhost:8001代表一个不同的节点

写一个启动Server服务的方法

`server.go`

```go
// 启动一个节点服务
func startServer(nodeID string, minerAdd string) {
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
```

针对不同的命令要采取不同的处理方式(上面已经讲了具体命令对应的实现)，所以需要实现一个命令解析器

`server.go`

```go
// 处理节点连接中的数据请求
func handleConnection(conn net.Conn, bc *pbcc.BlockChain) {
	// 读取客户端发送过来的所有的数据
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("收到的消息类型是:%s\n", request[:conf.COMMANDLENGTH])
	//version
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
```

到目前为止，服务器和通信已经搭建完毕，项目此时应该会有一个报错，因为获取区块链对象的时候我们传了节点地址作为参数，之前我们搭建的区块链是没有这个参数的，因为现在分节点了，那么在读取区块链数据的时候肯定是根据本节点的数据来读取，那么我们需要对之前的一些接口进行改造，接下来我们先从cli改动

## 10、命令行参数改动

之前的命令都是默认当前项目下的数据库，没有按照节点划分，下面我们所有的命令都要传入节点地址，这样命令就会根据节点去操作不同节点的数据库

接下来对所有的命令先进行补充节点参数

`cli.go`

```go
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
```

下面就来对之前的每个命令的处理逻辑进行改动

`cli_send.go`

```go
//转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {
	blockchain := pbcc.GetBlockchainObject(nodeID)
	utxoSet := &pbcc.UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
	defer blockchain.DB.Close()
	if mineNow {
		blockchain.MineNewBlock(from, to, amount, nodeID)
		//转账成功以后，需要更新一下
		utxoSet.Update()
	} else {
		// 把交易发送到矿工节点去进行验证
		fmt.Println("由矿工节点处理......")
		value, _ := strconv.Atoi(amount[0])
		tx := pbcc.NewSimpleTransaction(from[0], to[0], int64(value), utxoSet, []*pbcc.Transaction{}, nodeID)
		// 向全节点发送一下
		server.SendTx(server.KnowNodes[0], tx)
	}
}
```

`cli_print.go`

```go
// 打印节点的区块链信息
func (cli *CLI) printChains(nodeID string) {
	bc := pbcc.GetBlockchainObject(nodeID)
	if bc == nil {
		fmt.Println("没有区块可以打印")
		os.Exit(1)
	}
	defer bc.DB.Close()
	bc.PrintChains()
}
```

`cli_createblockchain.go`

```go
// 创建区块链
func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	pbcc.CreateBlockChainWithGenesisBlock(address, nodeID)
	bc := pbcc.GetBlockchainObject(nodeID)
	defer bc.DB.Close()
	if bc != nil {
		utxoSet := &pbcc.UTXOSet{bc}
		utxoSet.ResetUTXOSet()
	}
}
```

`cli_getbalance.go`

```go
//查询余额
func (cli *CLI) getBalance(address string, nodeID string) {
	fmt.Println("查询余额：", address)
	bc := pbcc.GetBlockchainObject(nodeID)
	if bc == nil {
		fmt.Println("数据库不存在，无法查询")
		os.Exit(1)
	}
	defer bc.DB.Close()
	utxoSet := &pbcc.UTXOSet{bc}
	utxoSet.ResetUTXOSet()
	balance := utxoSet.GetBalance(address)
	fmt.Printf("%s,一共有%d个Token\n", address, balance)
}
```

`cli_createwallet.go`

```go
// 创建一个新钱包地址
func (cli *CLI) createWallet(nodeID string) {
	wallets := wallet.NewWallets(nodeID)
	wallets.CreateNewWallet(nodeID)
}
```

`cli_getaddresslist.go`

```go
// 打印所有钱包地址
func (cli *CLI) addressLists(nodeID string) {
	fmt.Println("打印所有的钱包地址")
	//获取
	Wallets := wallet.NewWallets(nodeID)
	for address, _ := range Wallets.WalletsMap {
		fmt.Println("address:", address)
	}
}
```

`cli_utxoset.go`

```go
// 输出UXTO
func (cli *CLI) TestMethod(nodeID string) {
	blockchain := pbcc.GetBlockchainObject(nodeID)
	unSpentOutputMap := blockchain.FindUnSpentOutputMap()
	fmt.Println(unSpentOutputMap)
	for key, value := range unSpentOutputMap {
		fmt.Println(key)
		for _, utxo := range value.UTXOS {
			fmt.Println("金额：", utxo.Output.Value)
			fmt.Printf("地址：%v\n", utxo.Output.PubKeyHash)
			fmt.Println("---------------------")
		}
	}
	utxoSet := &pbcc.UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}
```

接下来新建cli_server.go，用于启动节点服务

`cli_server.go`

```go
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
```

关于命令行函数的改动已经完成了，这个时候项目已经是无法运行，因为我们在调用区块链的接口的时候都传入了节点的参数，接下来我们也需要对之前的接口进行改造

## 11、原接口改造

主要是添加上节点的信息参数，每个节点在进行操作的时候都是根据节点去获取自己的数据信息

`blockchain.go`

```go
//创建区块链，带有创世区块
func CreateBlockChainWithGenesisBlock(address string, nodeID string) {
	/*
		格式化数据库的名字
			1.修改数据库的名字："blockchain_%s.db"
			2.根据节点生成数据库的名字

	*/
	DBNAME := fmt.Sprintf(conf.DBNAME, nodeID)
	if dbExists(DBNAME) {
		fmt.Println("数据库已经存在")
		return
	}
	fmt.Println("创建创世区块：")
	//数据库不存在，说明第一次创建，然后存入到数据库中
	//创建创世区块
	//先创建coinbase交易
	txCoinBase := NewCoinBaseTransaction(address)
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})
	//打开数据库
	db, err := bolt.Open(DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(conf.BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误。。。")
			}
			//存储最新区块的hash
			b.Put([]byte("l"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// 获取最新的区块链
func GetBlockchainObject(nodeID string) *BlockChain {
	DBNAME := fmt.Sprintf(conf.DBNAME, nodeID)
	/*
		1.如果数据库不存在，直接返回nil
		2.读取数据库
	*/
	if !dbExists(DBNAME) {
		fmt.Println("数据库不存在，无法获取区块链")
		return nil
	}
	db, err := bolt.Open(DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var blockchain *BlockChain
	//读取数据库
	err = db.View(func(tx *bolt.Tx) error {
		//打开表
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			//读取最后一个hash
			hash := b.Get([]byte("l"))
			//创建blockchain
			blockchain = &BlockChain{hash, db}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return blockchain
}

//提供一个方法，用于判断数据库是否存在
func dbExists(DBName string) bool {
	if _, err := os.Stat(DBName); os.IsNotExist(err) {
		return false
	}
	return true
}

//挖掘新的区块 有交易的时候就会调用
func (bc *BlockChain) MineNewBlock(from, to, amount []string, nodeID string) {
	//新建交易
	//新建区块
	//将区块存入到数据库
	var txs []*Transaction
	//奖励
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)
	utxoSet := &UTXOSet{bc}
	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		tx := NewSimpleTransaction(from[i], to[i], amountInt, utxoSet, txs, nodeID)
		txs = append(txs, tx)
	}

	var block *Block    //数据库中的最后一个block
	var newBlock *Block //要创建的新的block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) //数据库中的最后一个block
		}
		return nil
	})
	//在建立新区块钱，对txs进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx, _txs) != true {
			log.Panic("签名验证失败。。")
		}
		_txs = append(_txs, tx)
	}

	newBlock = NewBlock(txs, block.Hash, block.Height+1)
	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(conf.BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("l"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
}
```

`wallets.go`

```go
// 获取钱包集，如果数据库有就从数据库获取，如果没有就创建
func NewWallets(nodeID string) *Wallets {
	walletFile := fmt.Sprintf(conf.WalletFile, nodeID)
	//判断钱包文件是否存在
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		fmt.Println("文件不存在")
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets
	}
	//否则读取文件中的数据
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	return &wallets
}

//钱包集创建一个新钱包
func (ws *Wallets) CreateNewWallet(nodeID string) {
	wallet := NewWallet()
	fmt.Printf("创建钱包地址：%s\n", wallet.GetAddress())
	ws.WalletsMap[string(wallet.GetAddress())] = wallet
	//将钱包保存
	ws.SaveWallets(nodeID)
}

/*
要让数据对象能在网络上传输或存储，我们需要进行编码和解码。
现在比较流行的编码方式有JSON,XML等。然而，Go在gob包中为我们提供了另一种方式，该方式编解码效率高于JSON。
gob是Golang包自带的一个数据结构序列化的编码/解码工具
*/
func (ws *Wallets) SaveWallets(nodeID string) {
	walletFile := fmt.Sprintf(conf.WalletFile, nodeID)
	var content bytes.Buffer
	//注册的目的，为了可以序列化任何类型，wallet结构体中有接口类型。将接口进行注册
	gob.Register(elliptic.P256()) //gob是Golang包自带的一个数据结构序列化的编码/解码工具
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	//将序列化后的数据写入到文件，原来的文件中的内容会被覆盖掉
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
```

修改一下钱包和数据库的名字，在输出的文件的是，让文件名带有节点信息

`config.go`

```go
const DBNAME = "blockchain_%s.db"  //数据库名
const WalletFile = "Wallets_%s.dat"
```

`transaction.go`

```go
func NewSimpleTransaction(from, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction, nodeID string) *Transaction {
	var txInputs []*TXInput
	var txOutputs []*TXOuput
	balance, spendableUTXO := utxoSet.FindSpendableUTXOs(from, amount, txs)

	//获取钱包
	wallets := wallet.NewWallets(nodeID)
	wallet := wallets.WalletsMap[from]

	for txID, indexArray := range spendableUTXO {
		txIDBytes, _ := hex.DecodeString(txID)
		for _, index := range indexArray {
			txInput := &TXInput{txIDBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput1 := NewTXOuput(amount, to)
	txOutputs = append(txOutputs, txOutput1)

	//找零
	txOutput2 := NewTXOuput(balance-amount, from)
	txOutputs = append(txOutputs, txOutput2)
	
	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.SetTxID()

	//进行签名
	utxoSet.BlockChain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}
```

终于我们项目可以正常运行了，接下来就是P2P网络的测试了

## 12、测试

接下来先举一个例子来更好的了解一下这个网络通信的过程

假设现在的情况是这样的：

A节点(中心节点)，拥有3个区块的区块链

B节点(钱包节点)，拥有1个区块的区块链

C节点(挖矿节点)，拥有1个区块的区块链

很明显，B节点需要向A节点请求2个区块更新到自己的区块链上。那么，实际的代码逻辑是怎样处理的？



中心节点与钱包节点的同步逻辑

A和B都是既可以充当服务端，也可以充当客户端。

> A.StartServer 等待接收其他节点发来的消息

> B.StartServer 启动同步服务

> B != 中心节点，向中心节点发请求:B.sendVersion(A, B.blc)

> A.Handle(B.Versin) :A收到B的Version消息 4.1 A.blc.Height > B.blc.Height(3>1) A.sendVersion(B, A.blc)

> B.Handle(A.Version):B收到A的Version消息 5.1 B.blc.Height > A.blc.Height(1<3) B向A请求其所有的区块哈希:B.sendGetBlocks(B)

> A.Handle(B.GetBlocks) A将其所有的区块哈希返回给B:A.sendInv(B, "block",blockHashes)

> B.Handle(A.Inv) B收到A的Inv消息 7.1取第一个哈希，向A发送一个消息请求该哈希对应的区块:B.sendGetData(A, blockHash) 7.2在收到的blockHashes去掉请求的blockHash后，缓存到一个数组unslovedHashes中

> A.Handle(B.GetData) A收到B的GetData请求，发现是在请求一个区块 8.1 A取出对应得区块并发送给B:A.sendBlock(B, block)

> B.Handle(A.Block) B收到A的一个Block 9.1 B判断该Block自己是否拥有，如果没有加入自己的区块链 9.2 len(unslovedHashes) != 0，如果还有区块未处理，继续发送GetData消息，相当于回7.1:B.sendGetData(A,unslovedHashes[0]) 9.3 len(unslovedHashes) == 0,所有A的区块处理完毕，重置UTXO数据库

> 大功告成

![image-20220718173544322](/Users/yunphant/Desktop/%E5%AD%A6%E8%80%8C%E5%AE%9E%E4%B9%A0%E6%97%B6/image-20220718173544322.png)

矿节点参与的同步逻辑

上面的同步并没有矿工挖矿的工作，那么由矿工节点参与挖矿时的同步逻辑又是怎样的呢？

> A.StartServer 等待接收其他节点发来的消息

> C.StartServer 启动同步服务，并指定自己为挖矿节点，指定挖矿奖励接收地址

> C != 中心节点，向中心节点发请求:C.sendVersion(A, C.blc)

> A.Handle(C.Version),该步骤如果有更新同上面的分析相同

> B.Send(B, C, amount) B给C的地址转账形成一笔交易 5.1 B.sendTx(A, tx) B节点将该交易tx转发给主节点做处理 5.2 A.Handle(B.tx) A节点将其信息分装到Inv发送给其他节点:A.SendInv(others, txInv)

> C.Handle(A.txInv),C收到转发的交易将其放到交易缓冲池memTxPool，当memTxPool内Tx达到一定数量就进行打包挖矿产生新区块并发送给其他节点：C.sendBlock(others, blockData)

> A(B).HandleBlock(C. blockData) A和B都会收到C产生的新区块并添加到自己的区块链上

> 大功告成



接下来就正式的开始测试，确保项目可以正常跑起来以后，运行`go build main.go`编译一下

打开三个终端分别模拟三个节点

终端1充当全节点、终端2充当轻节点、终端3充当旷工节点

下面的先阐述一下总的测试用例，三个节点分别创建自己的区块链，然后再主节点中先发起一笔交易（指定发起的时候就挖矿），然后主节点启动节点服务，钱包节点和旷工启动服务，同步全节点的区块链，由钱包节点发起一笔交易（指定由旷工处理），然后验证三个节点的区块链数据

注意命令行startnode后加-miner 指定地址 的话说明这个是旷工节点、send后加-mine的话就是直接打包成区块，不加的话就是由旷工处理挖矿



我们搭建的是比较简单的网络，大家还是按照本教程的测试步骤来，否则会出现bug，要是有小伙伴热衷于改bug的话，那可以自己折腾折腾，我们后续回去分析以太坊的源码，所有这里就不去实现那么复杂的网络了

终端1

```
1.设置节点端口为8000
2.创建钱包
3.创建区块链

(base) yunphant@yunphantdeMacBook-Pro publicchain % export NODE_ID=8000
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createwallet
本节点的NODE_ID是:8000
文件不存在
创建钱包地址：1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createblockchain -address 1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB
本节点的NODE_ID是:8000
创建创世区块：
77682: 0000ed62233ff555606e33e8a924f32bd69dc0e44114dd6eae5f360e4a20cacb

这个时候项目下多了两个文件 8000节点的区块链和钱包
```

切换终端2

```
1.设置节点端口为8001
2.创建钱包
3.创建区块链

(base) yunphant@yunphantdeMacBook-Pro publicchain % export NODE_ID=8001
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createwallet
本节点的NODE_ID是:8001
文件不存在
创建钱包地址：1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createblockchain -address 1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc
本节点的NODE_ID是:8001
创建创世区块：
63955: 00001c6cfa5857d2c95504fc92ebff91428206d3f64d0646fafd850de34e2e93

这个时候项目下多了两个文件 8001节点的区块链和钱包
```

这样我们可以发现全节点和钱包节点创建并不是同一条链，接下来就要来开启通信来同步

切换终端1

```
1.对终端2创建的钱包地址进行一次转账
2.启动节点服务

(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main send -from '["1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB"]' -to '["1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc"]' -amount '["5"]' -mine
本节点的NODE_ID是:8000
看下默认的mine true
5 ,未打包，转账花费： 10
4682: 00005e4e579859ede2c6f399ec5322996a71f3dc31e7c64a862de4c86a746ced
inputs的长度: 1 [0xc00007c140]
outsMaps, 5
outsMaps, 5
outsMap的长度: 1 map[17c02fd98dcf1a00e5c87aafa5a4e029b22c4fee346457aaa14593ac6214937c:0xc00020ca38]
0 =========================
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main startnode
本节点的NODE_ID是:8000
8000 
启动服务器:localhost:8000
nodeAddress:localhost:8000,minerAddress:


可以发现发起一笔交易以后，全节点里面是有两个区块了
```

切换终端2

```go
1.启动节点服务
2.control+c关闭链接
3.查询区块链数据
4.查询全节点钱包地址余额和本节点的钱包余额

(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main startnode
本节点的NODE_ID是:8001
8001 
启动服务器:localhost:8001
nodeAddress:localhost:8001,minerAddress:
主节点是:localhost:8000
节点localhost:8001向节点localhost:8000发送了version消息
收到的消息类型是:version
向节点地址为:localhost:8000的节点发送了GetBlock消息
收到的消息类型是:inv
节点localhost:8001向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 00005e4e579859ede2c6f399ec5322996a71f3dc31e7c64a862de4c86a746ced
节点localhost:8001向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 0000ed62233ff555606e33e8a924f32bd69dc0e44114dd6eae5f360e4a20cacb
已经没有要处理的区块了
^C
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main printchain
本节点的NODE_ID是:8001
第2个区块的信息:
        高度:1
        上一个区块的hash:0000ed62233ff555606e33e8a924f32bd69dc0e44114dd6eae5f360e4a20cacb
        当前的hash:00005e4e579859ede2c6f399ec5322996a71f3dc31e7c64a862de4c86a746ced
        交易:
                交易ID:0ff7f7f8146a2ca9bcd0c22cf8ab94f8a94ed67c39c466646bcdd8a93fd5e83f
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[171 175 218 38 66 160 156 192 244 184 101 72 179 199 220 77 192 70 138 3]
                交易ID:17c02fd98dcf1a00e5c87aafa5a4e029b22c4fee346457aaa14593ac6214937c
                Vins:
                        TxID:0ff7f7f8146a2ca9bcd0c22cf8ab94f8a94ed67c39c466646bcdd8a93fd5e83f
                        Vout:0
                        PublicKey:[25 116 197 230 168 26 175 247 49 138 66 60 58 191 147 201 16 178 45 240 157 175 53 106 13 9 53 141 119 120 159 98 82 125 43 76 120 166 123 111 65 13 159 107 125 99 131 182 181 42 189 241 222 169 71 158 114 45 95 203 246 220 127 193]
                Vouts:
                        value:5
                        PubKeyHash:[186 177 167 82 10 69 188 139 219 68 82 64 252 29 37 50 233 126 163 195]
                        value:5
                        PubKeyHash:[171 175 218 38 66 160 156 192 244 184 101 72 179 199 220 77 192 70 138 3]
        时间:2022-07-18 23:55:23
        次数:4682
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:0000ed62233ff555606e33e8a924f32bd69dc0e44114dd6eae5f360e4a20cacb
        交易:
                交易ID:31940224d27d02d6b50b24ba74f7fbc3879fd3007cacbeebc4f4c724ceeacb77
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[171 175 218 38 66 160 156 192 244 184 101 72 179 199 220 77 192 70 138 3]
        时间:2022-07-18 23:50:29
        次数:77682
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main getbalance -address 1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB
本节点的NODE_ID是:8001
查询余额： 1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB
1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB 余额： 5
1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB 余额： 10
1GeoDE1eYfXC2iBvsDPLcFyc5YLpnTghgB,一共有15个Token
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main getbalance -address 1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc
本节点的NODE_ID是:8001
查询余额： 1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc
1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc 余额： 5
1J29WGaq4tNo86yQSGj8Fz8EDLbRSvQXEc,一共有5个Token

由于在主节点发起了一笔交易，那么全节点的区块高度就比钱包节点高，那钱包节点就会向全节点请求区块同步，那么我们输出以后发现，钱包节点也已经完全把全节点的数据同步过来了
```

切换终端三

```
1.设置节点端口为8002
2.创建钱包
3.创建区块链
4.启动服务同步区块同时指定地址

(base) yunphant@yunphantdeMacBook-Pro publicchain % export NODE_ID=8002
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createwallet
本节点的NODE_ID是:8002
文件不存在
创建钱包地址：1Li5KidRd4ta7AhNpeXJEUEEd3UFhMmmxo
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main createblockchain -address 1Li5KidRd4ta7AhNpeXJEUEEd3UFhMmmxo
本节点的NODE_ID是:8002
创建创世区块：
71442: 000099994d3cdbb85240b4594334c4e0a2f801546acd40e2fa9ab0f8cd3ad101
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main startnode -miner 1Li5KidRd4ta7AhNpeXJEUEEd3UFhMmmxo
本节点的NODE_ID是:8002
8002 1Li5KidRd4ta7AhNpeXJEUEEd3UFhMmmxo
启动服务器:localhost:8002
nodeAddress:localhost:8002,minerAddress:1Li5KidRd4ta7AhNpeXJEUEEd3UFhMmmxo
主节点是:localhost:8000
节点localhost:8002向节点localhost:8000发送了version消息
收到的消息类型是:version
向节点地址为:localhost:8000的节点发送了GetBlock消息
收到的消息类型是:inv
节点localhost:8002向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 00005e4e579859ede2c6f399ec5322996a71f3dc31e7c64a862de4c86a746ced
节点localhost:8002向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 0000ed62233ff555606e33e8a924f32bd69dc0e44114dd6eae5f360e4a20cacb
已经没有要处理的区块了
```

切换终端2

```
1.前提：终端1和3的服务都是开着的
2.钱包节点发起一笔交易并指定由旷工打包区块（给旷工节点地址转账）
3.启动服务同步

(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main send -from '["1EGSkMkug3BcpqhCvro1ghsNs2q8Bias3N"]' -to '["13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7"]' -amount '["5"]'
本节点的NODE_ID是:8001
由矿工节点处理......
5 ,数据库，转账花费： 5
节点向节点localhost:8000发送了Tx消息
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main startnode
本节点的NODE_ID是:8001
8001 
启动服务器:localhost:8001
nodeAddress:localhost:8001,minerAddress:
主节点是:localhost:8000
节点localhost:8001向节点localhost:8000发送了version消息
收到的消息类型是:version
向节点地址为:localhost:8000的节点发送了GetBlock消息
收到的消息类型是:inv
节点localhost:8001向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 0000ca220e52907c2ee812d70e6dd682700a1e6a2d385d68ef0c1c78e15186ff
节点localhost:8001向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
节点localhost:8001向节点localhost:8000发送了GetData消息
收到的消息类型是:block
Recevied a new block!
Added block 0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
已经没有要处理的区块了

这个时候切换到终端1和终端3，能够看到，钱包节点给全节点发送了Tx消息，然后全节点给旷工节点发送了Tx消息，旷工节点进行挖矿打包交易，然后把最新的链同步给全节点
```

查询三个节点的区块链，比较信息是否一致

```
//全节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main printchain
本节点的NODE_ID是:8000
第3个区块的信息:
        高度:2
        上一个区块的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        当前的hash:0000ca220e52907c2ee812d70e6dd682700a1e6a2d385d68ef0c1c78e15186ff
        交易:
                交易ID:61eeb4a831f1c29ef8c43ece6eeeda2c8be116348ad56cc9d82c0d88e3e4aca3
                Vins:
                        TxID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                        Vout:0
                        PublicKey:[246 220 74 114 23 147 143 10 164 232 10 60 145 156 58 236 69 36 201 135 64 91 140 95 125 136 97 238 116 236 152 141 96 151 129 47 38 22 10 212 234 19 224 115 151 129 44 17 28 5 95 58 233 226 0 54 149 179 132 167 203 179 22 11]
                Vouts:
                        value:5
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
                        value:0
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                交易ID:ec86295a5de013072025de92fbd39371d5ca20ef5f61762945d96ca49476faf1
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
        时间:2022-07-20 09:54:13
        次数:25503
第2个区块的信息:
        高度:1
        上一个区块的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        当前的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        交易:
                交易ID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
                交易ID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                Vins:
                        TxID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                        Vout:0
                        PublicKey:[201 230 7 159 83 158 15 8 35 112 235 138 76 87 20 179 243 196 132 232 216 41 253 42 23 158 141 47 82 23 53 50 164 175 10 253 241 3 141 191 174 180 31 70 98 111 229 82 192 187 247 103 70 34 229 119 125 197 67 10 6 4 18 239]
                Vouts:
                        value:5
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                        value:5
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:45:52
        次数:75338
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        交易:
                交易ID:b19510a667049e5f962b09525cde38f77201890cb93b7a2d6eef22cf567611f9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:27:22
        次数:121422

//钱包节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main printchain
本节点的NODE_ID是:8001
第3个区块的信息:
        高度:2
        上一个区块的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        当前的hash:0000ca220e52907c2ee812d70e6dd682700a1e6a2d385d68ef0c1c78e15186ff
        交易:
                交易ID:61eeb4a831f1c29ef8c43ece6eeeda2c8be116348ad56cc9d82c0d88e3e4aca3
                Vins:
                        TxID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                        Vout:0
                        PublicKey:[246 220 74 114 23 147 143 10 164 232 10 60 145 156 58 236 69 36 201 135 64 91 140 95 125 136 97 238 116 236 152 141 96 151 129 47 38 22 10 212 234 19 224 115 151 129 44 17 28 5 95 58 233 226 0 54 149 179 132 167 203 179 22 11]
                Vouts:
                        value:5
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
                        value:0
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                交易ID:ec86295a5de013072025de92fbd39371d5ca20ef5f61762945d96ca49476faf1
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
        时间:2022-07-20 09:54:13
        次数:25503
第2个区块的信息:
        高度:1
        上一个区块的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        当前的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        交易:
                交易ID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
                交易ID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                Vins:
                        TxID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                        Vout:0
                        PublicKey:[201 230 7 159 83 158 15 8 35 112 235 138 76 87 20 179 243 196 132 232 216 41 253 42 23 158 141 47 82 23 53 50 164 175 10 253 241 3 141 191 174 180 31 70 98 111 229 82 192 187 247 103 70 34 229 119 125 197 67 10 6 4 18 239]
                Vouts:
                        value:5
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                        value:5
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:45:52
        次数:75338
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        交易:
                交易ID:b19510a667049e5f962b09525cde38f77201890cb93b7a2d6eef22cf567611f9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:27:22
        次数:121422

//旷工节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main printchain
本节点的NODE_ID是:8002
第3个区块的信息:
        高度:2
        上一个区块的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        当前的hash:0000ca220e52907c2ee812d70e6dd682700a1e6a2d385d68ef0c1c78e15186ff
        交易:
                交易ID:61eeb4a831f1c29ef8c43ece6eeeda2c8be116348ad56cc9d82c0d88e3e4aca3
                Vins:
                        TxID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                        Vout:0
                        PublicKey:[246 220 74 114 23 147 143 10 164 232 10 60 145 156 58 236 69 36 201 135 64 91 140 95 125 136 97 238 116 236 152 141 96 151 129 47 38 22 10 212 234 19 224 115 151 129 44 17 28 5 95 58 233 226 0 54 149 179 132 167 203 179 22 11]
                Vouts:
                        value:5
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
                        value:0
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                交易ID:ec86295a5de013072025de92fbd39371d5ca20ef5f61762945d96ca49476faf1
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[24 140 183 59 188 92 53 169 151 148 242 74 39 52 148 109 36 235 54 226]
        时间:2022-07-20 09:54:13
        次数:25503
第2个区块的信息:
        高度:1
        上一个区块的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        当前的hash:00009e9b3fa401ff71bda7a063f56c0d5f31997f53fc213866307551bf5421e6
        交易:
                交易ID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
                交易ID:174eb1223b665605c4f582bf4c72f2ff231ea17ff4f1a3bfff4e71064cdc01c9
                Vins:
                        TxID:476455aaa2055338df777e40bb601228aa20e8c872089a209983180d1dad07a9
                        Vout:0
                        PublicKey:[201 230 7 159 83 158 15 8 35 112 235 138 76 87 20 179 243 196 132 232 216 41 253 42 23 158 141 47 82 23 53 50 164 175 10 253 241 3 141 191 174 180 31 70 98 111 229 82 192 187 247 103 70 34 229 119 125 197 67 10 6 4 18 239]
                Vouts:
                        value:5
                        PubKeyHash:[145 133 94 26 63 160 199 136 151 195 214 74 67 214 199 65 89 32 94 22]
                        value:5
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:45:52
        次数:75338
第1个区块的信息:
        高度:0
        上一个区块的hash:0000000000000000000000000000000000000000000000000000000000000000
        当前的hash:0000b204e5ca846fd393946c45a17c5215aa3b72f304773ed8dca1f41d941231
        交易:
                交易ID:b19510a667049e5f962b09525cde38f77201890cb93b7a2d6eef22cf567611f9
                Vins:
                        TxID:
                        Vout:-1
                        PublicKey:[]
                Vouts:
                        value:10
                        PubKeyHash:[120 153 24 149 64 120 155 218 227 189 171 20 250 1 36 220 10 87 47 4]
        时间:2022-07-20 09:27:22
        次数:121422
```

输出钱包的余额，看是否一致，全节点的钱包余额应该为15，钱包节点为0，旷工节点为15

```
//全节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main getbalance -address 1BzfVKT9mKjLdJL3J7jSaWqspafmNvTCeY
本节点的NODE_ID是:8000
查询余额： 1BzfVKT9mKjLdJL3J7jSaWqspafmNvTCeY
1BzfVKT9mKjLdJL3J7jSaWqspafmNvTCeY 余额： 5
1BzfVKT9mKjLdJL3J7jSaWqspafmNvTCeY 余额： 10
1BzfVKT9mKjLdJL3J7jSaWqspafmNvTCeY,一共有15个Token

//钱包节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main getbalance -address 1EGSkMkug3BcpqhCvro1ghsNs2q8Bias3N
本节点的NODE_ID是:8001
查询余额： 1EGSkMkug3BcpqhCvro1ghsNs2q8Bias3N
1EGSkMkug3BcpqhCvro1ghsNs2q8Bias3N 余额： 0
1EGSkMkug3BcpqhCvro1ghsNs2q8Bias3N,一共有0个Token

//旷工节点
(base) yunphant@yunphantdeMacBook-Pro publicchain % ./main getbalance -address 13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7
本节点的NODE_ID是:8002
查询余额： 13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7
13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7 余额： 5
13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7 余额： 10
13Eonhx1m8NdLg9YombV1PSk6RkznQF5c7,一共有15个Token
```

可以发现，三个节点都完成同步以后，各自数据库里面的区块链都一致了，简单的P2P网络就搭建起来了

这里大家可以是试一下，如果钱包节点转账完以后不去同步全节点，持续转账会出现什么的BUG，然后把这个改一下



后续的话会一起来带着大家阅读以太坊的源码，该教程就是先浅浅的学习一下。



有什么问题的可以联系我



微信：`13721072141`
