package block

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/sharecc"
	"DAG-Exp/src/utils"
)

//Message Type List
const (
	MessageTypeInit = iota
	MessageTypeNodeIp
	MessageTypeHandshake
	MessageTypeConnect
	MessageTypeSyncTest

	MessageTypeCreateDag
	MessageTypeRequestData
	MessageTypeResponse
	MessageTypeAck

	MessageSubmit
	MessageVote
)

//Message m
type Message struct {
	Type      int //message type
	Operator  int //how to operate
	Signature []byte
	PublicKey []byte
	Body      interface{}
}


type HandMessage struct {
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

// 系统建立时的消息，无需改动
func HandshakeMessage(p2pListenPort string, account account.AccountInDb, address string) *Message {
	m := &Message{Type: MessageTypeHandshake /*, RoutePath: make([]int64, 0)*/}
	m.Body = HandMessage{
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


// 实验用的消息构造函数
func SyncMessage(syncMsg sharecc.SyncTest) *Message {
	m := &Message{
		Type:      MessageTypeSyncTest,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      syncMsg,
	}
	return m
}

// 广播 DAG Tag 的消息封装函数，所有的类型均从这一个入口
func DagTagMessage(tag dag.DagTag) *Message{
	var msgType int
	switch tag.TagType {
	case dag.TOPTAG:
		msgType = MessageTypeCreateDag
	case dag.REQUESTTAG:
		msgType = MessageTypeRequestData
	case dag.RESPONSETAG:
		msgType = MessageTypeResponse
	case dag.ACKTAG:
		msgType = MessageTypeAck
	default:
		utils.Log.Error("Mismatch The Message Type")
	}
	m := &Message{
		Type:      msgType,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      tag,
	}
	return m
}

func EpochMessage(tag dag.EpochTag) *Message{
	var msgType int
	msgType = MessageSubmit
	m := &Message{
		Type:      msgType,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      tag,
	}
	return m
}

func VoteMessage(vote dag.Vote) *Message{
	var msgType int
	msgType = MessageVote
	m := &Message{
		Type:      msgType,
		Operator:  0,
		Signature: nil,
		PublicKey: nil,
		Body:      vote,
	}
	return m
}

