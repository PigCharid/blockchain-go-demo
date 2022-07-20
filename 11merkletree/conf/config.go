package conf

//256位Hash里面前面至少有16个零
const TargetBit = 16

const DBNAME = "blockchain.db"  //数据库名
const BLOCKTABLENAME = "blocks" //表名

const Version = byte(0x00)   //版本
const AddressChecksumLen = 4 //校验和的长度

const WalletFile = "Wallets.dat" //钱包集

const UtxoTableName = "utxoTable" //UTXO的表名
