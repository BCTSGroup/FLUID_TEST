package block

import (
	"bfc/src/account"
	"bfc/src/account/wallet"
	"bfc/src/sharecc"
	"bfc/src/utils"
	"encoding/json"
)

//Message Type List
const (
	MessageTypeInit = iota
	MessageTypeNodeIp
	MessageTypeHandshake
	MessageTypeConnect
	MessageTypeTransaction
	MessageTypeSyncBlockJson
	MessageTypeRenovateBlock
	MessageTypeBlockJson
	MessageTypePrePrepare
	MessageTypePrepare
	MessageTypeCommit
	MessageTypeSubmitSharingData
	MessageTypeGetSharedData
)

const (
	OperatorDelete = iota
	OpetatorAdd
)

//Message m
type Message struct {
	Type      int //message type
	Operator  int //how to operate
	Signature []byte
	PublicKey []byte
	Body      interface{}
	//RoutePath []int64
}
type AllBlockInfo struct {
	Height int64
	Block  []Block
}

/*------------------3.6 bfc接口：查询指定交易-----------------------*/

type ContractQueryResponseData struct {
	Contract Transaction `json:"contract"`
}

type ContractQueryResponse struct {
	Code int                       `json:"code"`
	Data ContractQueryResponseData `json:"data"`
	Msg  string                    `json:"msg"`
}


type VoteRequest struct {
	VoterAddress     string
	CandidateAddress string //the Address of the one he want to vote
	Pledge           uint64 //the cost he want to take
}

//StageMessage s
type StageMessage struct {
	Height int64
	Hash   string
	Signer string
}

type HandMessage struct {
	Height  int64
	P2pPort string
	Account account.AccountInDb
	Address string
}

type ConnectMessage struct {
	SelfAccount account.AccountInDb
	Address     string
	Port        string
}

//把本地信息打包上传到请求新加入的节点
type NodeIpAndAccountMessage struct {
	Address      string
	NodeIPs      []utils.NODEIP
	NodeAccounts map[string]account.AccountInDb
}

//InitMessage wrap up initialize message
func InitMessage(nodeID int64) *Message {
	m := &Message{Type: MessageTypeInit /*, RoutePath: make([]int64, 0)*/}
	m.Body = nodeID
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func HandshakeMessage(height int64, p2pListenPort string, account account.AccountInDb, address string) *Message {
	m := &Message{Type: MessageTypeHandshake /*, RoutePath: make([]int64, 0)*/}
	m.Body = HandMessage{
		Height:  height,
		P2pPort: p2pListenPort,
		Account: account,
		Address: address,
	}
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func NodeIpMessage(ips []utils.NODEIP, accounts map[string]account.AccountInDb) *Message {
	m := &Message{Type: MessageTypeNodeIp /*, RoutePath: make([]int64, 0)*/}
	m.Body = NodeIpAndAccountMessage{
		NodeIPs:      ips,
		NodeAccounts: accounts,
		Address:      GMyAccountAddress,
	}
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func ConnectBasedOnNodeIpMessage(account account.AccountInDb, address string, port string) *Message {
	m := &Message{Type: MessageTypeConnect /*, RoutePath: make([]int64, 0)*/}
	m.Body = ConnectMessage{
		SelfAccount: account,
		Address:     address,
		Port:        port,
	}
	return m
}

func BlockJsonRenovateMessage(block []byte) *Message {

	//context hash get
	msgByte, _ := json.Marshal(block)
	msgHash, _ := utils.GetHashFromBytes(msgByte)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(msgHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypeRenovateBlock /*, RoutePath: make([]int64, 0)*/}
	m.Body = block
	m.Signature = signature
	m.PublicKey = pkeyByte
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

//BlockMessage wrap up block message
func BlockJsonMessage(block []byte) *Message {

	//context hash get
	msgByte, _ := json.Marshal(block)
	msgHash, _ := utils.GetHashFromBytes(msgByte)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(msgHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypeBlockJson /*, RoutePath: make([]int64, 0)*/}
	m.Body = block
	m.Signature = signature
	m.PublicKey = pkeyByte
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func SyncBlockJsonMessage(block []byte) *Message {

	//context hash get
	msgByte, _ := json.Marshal(block)
	msgHash, _ := utils.GetHashFromBytes(msgByte)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(msgHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypeSyncBlockJson /*, RoutePath: make([]int64, 0)*/}
	m.Body = block
	m.Signature = signature
	m.PublicKey = pkeyByte
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func BlockTransactionMessage(msg Transaction, operatorType int) *Message {

	//context hash get
	msgByte, _ := json.Marshal(msg)
	msgHash, _ := utils.GetHashFromBytes(msgByte)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(msgHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypeTransaction, Operator: OperatorDelete}
	m.Signature = signature
	m.PublicKey = pkeyByte
	m.Body = msg
	return m
}

func PbftPrePrepareMessage(block Block) *Message {

	// get block's hash field as block hash
	blockHash := []byte(block.Head.Hash)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(blockHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypePrePrepare /*, RoutePath: make([]int64, 0)*/}
	m.Signature = signature
	m.PublicKey = pkeyByte
	m.Body = block

	return m
}

func PbftPrepareMessage(block Block) *Message {

	//get the block hash
	blockHash := []byte(block.Head.Hash)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(blockHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypePrepare /*, RoutePath: make([]int64, 0)*/}
	m.Signature = signature
	m.PublicKey = pkeyByte
	m.Body = block
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

func PbftCommitMessage(block Block) *Message {

	//get the block hash
	blockHash := []byte(block.Head.Hash)
	//get the context hash signature
	signature := account.GetEccSignFromHashData(blockHash, wallet.PRIVATE_PATH)
	//get the public key
	pkeyAxis := &account.GAccount.MyWallet.PrivateKey.PublicKey
	//translate the public key in axis in byte
	pkeyByte := wallet.MarshalPubkey(pkeyAxis)

	m := &Message{Type: MessageTypeCommit /*, RoutePath: make([]int64, 0)*/}
	m.Body = block
	m.Signature = signature
	m.PublicKey = pkeyByte
	//m.RoutePath = append(m.RoutePath, nodeID)
	return m
}

// 当前端接口收到 Submit Sharing Data 时，一方面本地调用共享数据接收合约，另一方面将该消息转发给其他节点
func SubmitSharingDataMessage(submitDataRequest sharecc.SubmitRequest) *Message {
	m := &Message{
		Type:      MessageTypeSubmitSharingData,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      submitDataRequest,
	}
	return m
}

func GetSharedDataRequestMessage(getRequest sharecc.GetSharedDataRequest) *Message{
	m := &Message{
		Type:      MessageTypeGetSharedData,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      getRequest,
	}
	return m
}
