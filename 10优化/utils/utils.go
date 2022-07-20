package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

//Json字符串转为[] string数组
func JSONToArray(jsonString string) []string {
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}
	return sArr
}

//字节数组反转
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
