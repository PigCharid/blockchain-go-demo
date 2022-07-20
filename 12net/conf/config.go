package conf

//256位Hash里面前面至少有16个零
const TargetBit = 16

const DBNAME = "blockchain_%s.db" //数据库名
const BLOCKTABLENAME = "blocks"   //表名

const Version = byte(0x00)   //版本
const AddressChecksumLen = 4 //校验和的长度

const WalletFile = "Wallets_%s.dat"

const UtxoTableName = "utxoTable" //UTXO的表名

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
