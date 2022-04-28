package block

import (
	"bfc/src/db_api"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"bfc/src/utils"
)

const (
	SLOT_TIME_INTERVAL     int64 = 3
	BLOCK_PRODUCE_INTERVAL int64 = 3
)

var GMyAccountAddress string



const (
	BLOCK_GENESIS = iota
	BLOCK_LIGHT
	BLOCK_REGISTER
	BLOCK_PRODUCER
)

//Block B
type BlockHead struct {
	Version       int64
	Height        int64
	TxCount       int
	Timestamp     int64
	Producer      string
	PrevBlockHash string
	Hash          string
	MerkleRoot	  []byte
}
type Block struct {
	Head           BlockHead
	//MTree          MerkleTree
	TransactionMap map[string]Transaction
}

func NewBlock() *Block {
	// 获取当前区块链高度
	GChainMemRWlock.RLock()
	height := GChainMem.Height
	prevHash := GChainMem.Hash
	GChainMemRWlock.RUnlock()

	// 填写区块数据
	GTransactionPoolRWlock.RLock()
	b := &Block{
		Head:BlockHead{
			//Hash:          "",
			Version:       1,
			Height:        height + 1,
			Timestamp:     time.Now().Unix(),
			Producer:      GMyAccountAddress,
			PrevBlockHash: prevHash,
		},
		TransactionMap: make(map[string]Transaction),
	}
	// 把交易池中的交易打包到 Block 中   (Ohh god! Ugly code...)
	//var keyInTransactionMap []string
	//for key_ := range GTransactionPool.TransactionMap {
	//	keyInTransactionMap = append(keyInTransactionMap, key_)
	//}
	//for i := 0; i < utils.MAX_TRANSACTION_IN_BLOCK && i < len(keyInTransactionMap); i++ {
	//	transactionID := GTransactionPool.TransactionMap[keyInTransactionMap[i]].TransactionID
	//	b.TransactionMap[transactionID] = GTransactionPool.TransactionMap[keyInTransactionMap[i]]
	//}

	//t1 := time.Now()
	//i := 0
	//for k, v := range GTransactionPool.TransactionMap {
	//	b.TransactionMap[k] = v
	//	i = i + 1
	//	if i >= 3000 { break }
	//}
	//t2 := time.Since(t1)
	//utils.Log.Critical("打包耗时 ： ", t2)

	Timer := time.NewTicker(time.Duration(1) * time.Millisecond)
	LOOP:
	for k, v := range GTransactionPool.TransactionMap {
		select {
		case <-Timer.C:
			break LOOP
		default:
			b.TransactionMap[k] = v
		}
	}
	utils.Log.Critical("打包交易：",len(b.TransactionMap))

	b.Head.TxCount = len(b.TransactionMap)
	// 计算 Merkle Tree
	b.CalculateMerkleTree()
	b.CalculateHash()
	GTransactionPoolRWlock.RUnlock()

	return b
}

//GetHash return hash of block
func (b *Block) GetHash() string {
	return b.Head.Hash
}

//GetPrevBlockHash return previous block hash of block
func (b *Block) GetPrevBlockHash() string {
	return b.Head.PrevBlockHash
}

//GetHeight return height of block
func (b *Block) GetHeight() int64 {
	return b.Head.Height
}

//GetTransactions return transactions of block
func (b *Block) GetTransactions() map[string]Transaction {
	return b.TransactionMap
}

//GetTimestamp return timestamp of block
func (b *Block) GetTimestamp() int64 {
	return b.Head.Timestamp
}

//GetForger return forger of block
func (b *Block) GetProducer() string {
	return b.Head.Producer
}

//CalculateHash calculate block hash value
func (b *Block) CalculateHash() {
	//hash := sha256.Sum256(buff.Bytes())
	b.Head.Hash = ""
	jsonBytes, _ := json.Marshal(b)
	hash := sha256.Sum256(jsonBytes)
	b.Head.Hash = hex.EncodeToString(hash[:])
}

//the received block should be the next one by the local height
func ValidateBlock(block *Block) bool {

	GChainMemRWlock.RLock()
	defer GChainMemRWlock.RUnlock()

	utils.Log.Infof("Block Height %d, Local height %d ", block.Head.Height, GChainMem.Height)
	utils.Log.Infof("Previous Hash  %s, Local Hash  %s", block.Head.PrevBlockHash, GChainMem.Hash)
	return GChainMem.Height+1 == block.Head.Height && block.Head.PrevBlockHash == GChainMem.Hash
}

//to renovate the situation when the local height is higher than the seed node
func ValidateRenovateBlock(block *Block) bool {

	if block.Head.Height <= 1 {
		return true
	}

	GChainMemRWlock.RLock()
	defer GChainMemRWlock.RUnlock()

	tmpVerificationBlock := Block{}
	verificationBlockByte, err := db_api.Db_getBlock(block.Head.Height - 1)
	returnValue := false
	if err != nil {
		returnValue = false
	} else {
		utils.Log.Infof("Block Height %d, Local height %d ", block.Head.Height, GChainMem.Height)
		utils.Log.Infof("Previous Hash  %s, Local Hash  %s", block.Head.PrevBlockHash, GChainMem.Hash)
		err = json.Unmarshal(verificationBlockByte, &tmpVerificationBlock)
		if err != nil {
			utils.Log.Errorf("unmarshal block failed, error: %s", err.Error())
			returnValue = false
		} else {
			if block.Head.PrevBlockHash == tmpVerificationBlock.Head.Hash {
				returnValue = true
			} else {
				returnValue = false
			}
		}
	}
	return returnValue

}

/*
	预执行交易内容，检查是否符合交易执行条件
	- 在 PBFT 的 PRE-PREPARE 阶段执行，needDelete为true，表示删除新打包的区块中无效的交易并重新计算hash
	- 在 PBFT 的 COMMIT 阶段执行，needDelete为false，表示检验是否存在无效交易，存在则返回error
*/
func PreExecuteBlock(block *Block, needDelete bool) error {
	// TODO 只验证系统模型的可行性，可暂时省略该步骤
	return nil
}