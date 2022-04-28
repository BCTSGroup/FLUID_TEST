package sharecc
// 请求类型宏定义
const (
	JOIN = iota
	SUBMIT
	GET_BLOCKHEAD
	GET_DATA
)
/******************************************************************************
							① 请求通用结构体
******************************************************************************/
// 接入请求
type JoinRequest struct {
	MsgType int 		`json:"msg_type"`
	From    string  	`json:"from"`
	Payload JoinPayload `json:"payload"`
}
type JoinPayload struct {
	ClassY00 []string	 `json:"class_y00"`
	ClassY01 []string	 `json:"class_y01"`
	ClassY10 []string	 `json:"class_y10"`
	ClassY11 []string	 `json:"class_y11"`
	AuthInfo interface{} `json:"auth_info"`
}

// 数据提交请求
type SubmitRequest struct {
	MsgType int 			`json:"msg_type"`
	From    string  		`json:"from"`
	Payload SubmitPayload 	`json:"payload"`
}
type SubmitPayload struct {
	Data 		[]ValidData `json:"data"`
	BlockHash 	string 		`json:"block_hash"`
}
type ValidData struct {
	DataType    string		`json:"data_type"`
	Key   		string		`json:"key"`
	Value 		string		`json:"value"`
	TimeValid   string		`json:"time_valid"`
	AuthInfo	interface{}	`json:"auth_info"`
}
type TxSharingData struct {
	// 打包进交易中的共享数据
	DataType    string
	Key   		string
	Value 		string
	TimeValid   string
}

// 请求共享数据结构体
type GetSharedDataRequest struct {
	MsgType 	int 					`json:"msg_type"`
	From    	string  				`json:"from"`
	Payload 	GetSharedDataPayload 	`json:"payload"`
	TimeStamp 	int64					`json:"time_stamp"`
}
type GetSharedDataPayload struct {
	DataType	string	`json:"data_type"`
	Key			string	`json:"key"`
	Signature 	string	`json:"signature"`
}

type GetResponse struct {
	Code		int					`json:"code"`
	Data		GetResponsePayload	`json:"data"`
	Timestamp	int64				`json:"timestamp"`
}
type GetResponsePayload struct {
	DataType    string 	`json:"data_type"`
	Key		  	string	`json:"key"`
	Value 	  	string	`json:"value"`
	TimeValid   string  `json:"time_valid"`
	BlockHash	string	`json:"block_hash"`
	MerklePath	MerklePath	`json:"merkle_path"`
}

type MerklePath struct {
	// ProofSet 就是验证用的路径 其中ProofSet[0] 就是需要验证的数据的 []byte 形式
	ProofSet  	[][]byte	`json:"proof_set"`
	ProofIndex 	uint64		`json:"proof_index"`
	NumLeaves	uint64		`json:"num_leaves"`
}

type DataSharingRecord struct {
	Requester  string
	RequestKey string
	Timestamp  int64
}
// 共享数据数据库中存储的数据
type SharedData struct {
	DataType    string
	Value 		string
	Level 		string
	ValidTime   string
	BlockHash	string
}

