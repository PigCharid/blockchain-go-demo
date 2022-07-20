package server

import "publicchain/pbcc"

//存储节点全局变量
var KnowNodes = []string{"localhost:8000"}            //localhost:8000 主节点的地址
var NodeAddress string                                //全局变量，节点地址
var TransactionArray [][]byte                         // 存储hash值
var MinerAddress string                               //旷工地址
var MemoryTxPool = make(map[string]*pbcc.Transaction) //交易池存储交易
