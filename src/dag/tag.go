package dag

import (
	"DAG-Exp/src/utils"
	"encoding/json"
	"sync"
)


const (
	TOPTAG = iota
	REQUESTTAG
	RESPONSETAG
	ACKTAG
	EPOCH
)



type DagTag struct {
	TimeStamp   int64           `json:"timeStamp"`
	Depth       int				`json:"depth"`
	Hash 		string			`json:"hash"`
	PrevHash 	[]string			`json:"prevHash"`
	Path        string			`json:"path"`
	TagType 	int				`json:"tagType"`
	Body 		interface{} 	`json:"body"`
	Miner 		string			`json:"miner"`
	Signature 	string			`json:"signature"`
}
/*

	1. 通过底层区块链发布共享的权限信息，并派生 TopTag
	2. 更新权限时也是先由底层区块链登记权限信息，再派生 TopTag，但区块链只负责共识记录，不部署权限控制合约

*/
type TopTagBody struct {
	PolicyVersion 	int
	PolicyAbstract 	string	// 权限摘要，详细的权限配置信息将存放在底层的区块链中
	PermissionsList map[string]string // [地址，权限] 由于是课题实验，用最简单的 string 来描述权限即可
}

/*

共享数据网络，数据资源在参与数据共享的所有节点那里都是明牌的，
意味着每个节点都知道哪个节点上有什么类型的数据，但没法知道数据
的具体内容。访问控制即是依据节点对共享数据的需求，凭借个人权限
向目标地址索取访问权限的过程。

*/
type RequestTagBody struct {
	FromAddress string
	ToAddress	string

	ReqInfo 	string
	Token 		string

	ReqHash		string
}

type ResponseTagBody struct {
	FromAddress string
	ToAddress	string

	ExecResult 	bool
	InfoURL		string
	Code 		string

	ReqHash		string
	PrevHash	[]string
}

type AckTagBody struct {
	FromAddress string
	ToAddress	string

	Pass 		bool
	ReqHash		string
	RespHash	string
	PrevHash	[]string
}

type EpochTag struct {
	TimeStamp 		int64 `json:"timeStamp"`
	Hash 			string `json:"hash"`
	PrevHashList 	[]string `json:"prev_hash_list"`
	Checksum 		string `json:"checksum"`
	Signature 		string `json:"signature"`
}

type Vote struct {
	EpochHash string `json:"epoch_hash"`
	TimeStamp int64 `json:"timeStamp"`
	Pass      int   `json:"pass"`
	Latency   int64   `json:"latency"`
}
var VOTELOCK sync.RWMutex
var VSLOCK sync.RWMutex
var VOTE map[string]int
var VoteLatency map[string]int64
var VoteLatencyResult map[string][]int64
func init (){
	VOTE = make(map[string]int)
	VoteLatency = make(map[string]int64)
	VoteLatencyResult = make(map[string][]int64)
}

func CalculateTagHash(tag DagTag) string{
	bTag, _ := json.Marshal(tag)
	bHash, _ := utils.GetHashFromBytes(bTag)
	return string(utils.Base58Encode(bHash))
}

func CalculateEpochTagHash(tag EpochTag) string{
	bTag, _ := json.Marshal(tag)
	bHash, _ := utils.GetHashFromBytes(bTag)
	return string(utils.Base58Encode(bHash))
}