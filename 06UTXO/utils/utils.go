package utils

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"publicchain/conf"
)

//将int64转换为bytes
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

//判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(conf.DBNAME); os.IsNotExist(err) {
		return false
	}
	return true
}
