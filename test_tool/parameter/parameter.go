package parameter


// Http 返回结构
type GetNewBlockResp struct {
	Height int64
	Block Block
}
// 区块结构
type Block struct {
	Version       int64
	Height        int64
	Timestamp     int64
	Producer      string
	PrevBlockHash string
	Hash          string
	TransactionMap map[string]Transaction
}
// 交易结构
type Transaction struct {
	TransactionID string
	TimeStamp     int64
	TransactionType int
	TransactionBody interface {}
}

type DAGTagCount struct {
	N    int `json:"n"`
	NA   int `json:"na"`
}

