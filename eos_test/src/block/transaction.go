package block

import (
	"encoding/json"
	"sync"
)

//全局交易池和交易池锁
var GTransactionPool TransactionPool
var GTransactionPoolRWlock sync.RWMutex

//交易的类型定义
const (
	TX_JOIN = iota //0
	TX_SUBMIT
	TX_SHARE
	TX_TEST
	TX_REQUEST
	TX_RESPONSE
)

type P2pBroadcastToSuperNodeTransactionCallback func(msg Transaction, operatorType int)
var P2PBroadcastToSuperCallback P2pBroadcastToSuperNodeTransactionCallback

const PLEDGE_TO_VOTE = 1

// 通用交易类型
type Transaction struct {
	TransactionID string
	TimeStamp     int64
	TransactionType int
	TransactionBody interface {}
}


func (tx Transaction) Serialize() []byte {
	bTx,_ := json.Marshal(tx.TransactionBody)
	return bTx
}



// DPOS 投票
//type TransactionVote struct {
//	AccountFrom string
//	AccountTo   string
//	Pledge      uint64
//}
//type PushTransactionResponse struct {
//	ResponseCode string
//	ExtraInfo    string
//}

//func (tx *Transaction) ProcessVoteTransaction() int {
//
//	map2 := tx.TransactionBody.(map[string]interface{})
//
//	msg := VoteRequest{
//		VoterAddress:     map2["VoterAddress"].(string),
//		CandidateAddress: map2["CandidateAddress"].(string),
//		Pledge:           uint64(map2["Pledge"].(float64)),
//	}
//
//	if msg.VoterAddress != msg.CandidateAddress {
//		utils.Log.Error("Processing VoteTransaction  :  Voting other!", msg.VoterAddress, msg.CandidateAddress)
//
//		CandidateAccount := account.AccountInDb{}
//		CandidateAccount, err := account.GetAccountAsStructureFromDB(msg.CandidateAddress)
//		if err != nil {
//			utils.Log.Errorf("Processing VoteTransaction  :  read candidate from db failed %s", err.Error())
//		}
//		VoterAccount := account.AccountInDb{}
//		VoterAccount, err = account.GetAccountAsStructureFromDB(msg.VoterAddress)
//		if err != nil {
//			utils.Log.Errorf("read voter from db failed  %s", err.Error())
//		}
//
//		CandidateAccount.Votes += msg.Pledge / PLEDGE_TO_VOTE
//		VoterAccount.Pledge += msg.Pledge
//		VoterAccount.Balance -= float64(msg.Pledge)
//
//		jsonVoterAccount, err := json.Marshal(VoterAccount)
//
//		if err != nil {
//			utils.Log.Errorf("VoterAccount serilized failed: %s", err.Error())
//		}
//
//		err = db_api.Db_insertAccount(jsonVoterAccount, msg.VoterAddress)
//
//		if err != nil {
//			utils.Log.Errorf("insert account failed %s", err.Error())
//		}
//
//		jsonCandidateAccount, err := json.Marshal(CandidateAccount)
//
//		if err != nil {
//			utils.Log.Errorf("VoterAccount serilized failed: %s", err.Error())
//		}
//
//		err = db_api.Db_insertAccount(jsonCandidateAccount, msg.CandidateAddress)
//
//		if err != nil {
//			utils.Log.Errorf("insert account failed %s", err.Error())
//		}
//	} else {
//		utils.Log.Debug("Voting self!", msg.VoterAddress, msg.CandidateAddress)
//
//		VoterAccount := account.AccountInDb{}
//		VoterAccount, err := account.GetAccountAsStructureFromDB(msg.VoterAddress)
//		if err != nil {
//			utils.Log.Errorf("read voter from db failed %s", err.Error())
//		}
//
//		VoterAccount.Votes += msg.Pledge / PLEDGE_TO_VOTE
//		VoterAccount.Pledge += msg.Pledge
//		VoterAccount.Balance -= float64(msg.Pledge)
//		jsonVoterAccount, err := json.Marshal(VoterAccount)
//
//		if err != nil {
//			utils.Log.Errorf("VoterAccount serilized failed: %s", err.Error())
//		}
//
//		err = db_api.Db_insertAccount(jsonVoterAccount, msg.VoterAddress)
//
//		if err != nil {
//			utils.Log.Errorf("insert account failed %s", err.Error())
//		}
//	}
//
//	return SUCCESS
//}
