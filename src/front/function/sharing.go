package function

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/db_api"
	Network "DAG-Exp/src/network"
	"DAG-Exp/src/sharecc"
	"DAG-Exp/src/utils"
	"encoding/json"
	"net/http"
)

// 请求反馈消息结构
const (
	SUCCESS = iota
	FAIL
)
type response struct {
	Status int
	Prompt interface{}
}

func HandleSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 解析 Request Body
	rDecoder := json.NewDecoder(r.Body)
	var msg sharecc.SyncTest
	err := rDecoder.Decode(&msg)
	if err != nil {
		RespondWithJSON(w, r, http.StatusBadRequest, r.Body)
		utils.Log.Errorf("Decode failed.[ %s ]", err)
		return
	}
	// 交易消息广播给其他节点
	Network.BroadcastSyncTestMsg(msg)
	// 返回请求提交成功反馈信息
	successResp := response{
		Status: SUCCESS,
		Prompt: "Request submitted successfully.",
	}
	RespondWithJSON(w, r, http.StatusOK, successResp)
	return
}

// 获取本地账户信息
func HandleGetLocalAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//var NodeAccount []db_struct.AccountInDb

	type AccountToReturn struct {
		PublicKey string
		//PrivateKey   string
		Balance      float64
		Votes        uint64
		Pledge       uint64
		NodeType     int
	}
	NodeAccounts := make(map[string]AccountToReturn)
	NodeAddresses, _ := db_api.Db_getAllAddressAsSlice()
	var NodeAccount account.AccountInDb
	for _, address := range NodeAddresses {
		NodeAccount, _ = account.GetAccountAsStructureFromDB(address)
		//NodeAccounts = append(NodeAccounts,NodeAccount)

		NodeAccounts[address] = AccountToReturn{
			PublicKey:    string(utils.Base58Encode(NodeAccount.PublicKey)),
			Balance:      NodeAccount.Balance,
		}
	}
	RespondWithJSON(w, r, http.StatusOK, NodeAccounts)

}