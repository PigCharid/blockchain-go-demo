package pbcc

import (
	"crypto/sha256"
	"math"
)

//默克尔树
type MerkleTree struct {
	RootNode *MerkleNode //根节点
}

//默克尔树节点
type MerkleNode struct {
	LeftNode  *MerkleNode //左节点
	RightNode *MerkleNode //右节点
	Data      []byte      //节点数据
}

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
