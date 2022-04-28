package block

import (
	"bfc/src/db_api"
	"bfc/src/sharecc"
	"bfc/src/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/merkletree"
	"sort"
)
//
////默克尔树
//type MerkleTree struct {
//	//根节点
//	RootNode *MerkleNode
//}
//
////默克尔树节点
//type MerkleNode struct {
//	//做节点
//	Left *MerkleNode
//	//右节点
//	Right *MerkleNode
//	//节点数据
//	Data []byte
//}
//
////新建节点
//func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
//
//	mNode := MerkleNode{}
//
//	if left == nil && right == nil {
//		hash := sha256.Sum256(data)
//		mNode.Data = hash[:]
//	} else {
//		prevHashes := append(left.Data, right.Data...)
//		hash := sha256.Sum256(prevHashes)
//		mNode.Data = hash[:]
//	}
//
//	mNode.Left = left
//	mNode.Right = right
//
//	return &mNode
//}
//
//// 1 2 3 --> 1 2 3 3
////新建默克尔树
//func NewMerkleTree(datas [][]byte) *MerkleTree {
//
//	var nodes []*MerkleNode
//
//	//如果是奇数，添加最后一个交易哈希拼凑为偶数个交易
//	if len(datas) % 2 != 0 {
//		datas = append(datas, datas[len(datas)-1])
//	}
//
//	//将每一个交易哈希构造为默克尔树节点
//	for _, data := range datas {
//		node := NewMerkleNode(nil, nil, data)
//		nodes = append(nodes, node)
//	}
//
//	//将所有节点两两组合生成新节点，直到最后只有一个更节点
//	for i := 0; i < len(datas)/2; i++ {
//		var newLevel []*MerkleNode
//
//		for j := 0; j < len(nodes); j += 2 {
//			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
//			newLevel = append(newLevel, node)
//		}
//		nodes = newLevel
//	}
//
//	//取根节点返回
//	mTree := MerkleTree{nodes[0]}
//
//	return &mTree
//}
// 需要将Txs转换成[]byte
func (block *Block) CalculateMerkleTree()  {
	//默克尔树根节点表示交易哈希
	//var transactions [][]byte
	//utils.Log.Debug("计算 Merkle Tree")
	//for _, tx := range block.TransactionMap {
	//
	//	transactions = append(transactions, tx.Serialize())
	//}
	//mTree := NewMerkleTree(transactions)
	//
	//block.Head.MerkleRoot = mTree.RootNode.Data
	//block.MTree = *mTree

	// 计算 Merkle 时保证叶子节点的顺序是有序的
	var keys []string
	for k := range block.TransactionMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	mTree := merkletree.New(sha256.New())

	for _, k := range keys {
		mTree.Push(block.TransactionMap[k].Serialize())
	}

	block.Head.MerkleRoot = mTree.Root()

}
func GetProofBySharedData(data sharecc.TxSharingData, hash string) (path [][]byte, proofIndex uint64, numLeaves uint64) {
	// 1. 找到数据原始区块
	b := db_api.DB_GetBlockByHash(hash)
	var block Block
	json.Unmarshal(b, &block)

	// 2. 还原数据构成的交易 hash
	txBody, _ := json.Marshal(data)
	bTxHash, _ := utils.GetHashFromBytes(txBody)
	txHash := hex.EncodeToString(bTxHash)
	utils.Log.Critical("共享时的交易 Hash，数据 : ",txHash, string(txBody))
	// 3. 有序化区块中的 Tx Map
	var keys []string
	for k := range block.TransactionMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 4. 初始化 Merkle Tree 设置数据验证索引
	mTree := merkletree.New(sha256.New())
	for i, key := range keys {
		if key == txHash {
			utils.Log.Debug("找到对应的数据，完成设置Merkle proof index.")
			mTree.SetIndex(uint64(i))
			break
		}
	}

	// 5. 还原 Merkle Tree
	for _, key := range keys {
		var txMeshData sharecc.TxSharingData
		bTxMeshData, err := json.Marshal(block.TransactionMap[key].TransactionBody)
		// utils.Log.Critical("还原Merkle Tree(bTxMeshData) : ", string(bTxMeshData))
		if err != nil {
			utils.Log.Error(err)
		}
		err = json.Unmarshal(bTxMeshData, &txMeshData)
		if err != nil {
			utils.Log.Error(err)
		}
		bTx, _ := json.Marshal(txMeshData)
		// utils.Log.Critical("还原Merkle Tree(.push) : ", string(bTx))
		mTree.Push(bTx)
	}

	// 6. 获取验证路径
	_, path, proofIndex, numLeaves = mTree.Prove()


	return path, proofIndex, numLeaves
}
