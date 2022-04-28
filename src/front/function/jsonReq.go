package function

import (
	"DAG-Exp/src/dag"
)

type ReqCreateDAG struct {
	Depth 		int 				`json:"depth"`
	Hash 		string				`json:"hash"`
	PrevHash 	[]string				`json:"prevHash"`
	TagType 	int					`json:"tagType"`
	Body 		dag.TopTagBody 	`json:"body"`
	Miner 		string				`json:"miner"`
	Signature 	string				`json:"signature"`
}

type ReqRequestData struct {
	FromAddress string	`json:"fromAddress"`
	ToAddress	string	`json:"toAddress"`
	ReqInfo 	string	`json:"reqInfo"`
	Token 		string	`json:"token"`
	ReqHash		string	`json:"reqHash"` // 客户端发送的时候空着
}
