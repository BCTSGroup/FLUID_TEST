package function

import (
	"bfc/src/account"
	"bfc/src/block"
	"bfc/src/db_api"
	"bfc/src/test"
	"bfc/src/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// 查询合约
//func HandleContractQuery(w http.ResponseWriter, r *http.Request) {
//	// Read the contract ID
//	vars := mux.Vars(r)
//	contractId := vars["contractID"]
//	// Query the contract by the contract ID
//	jsonContractFound, err := func(contractID string) ([]byte, error) {
//		var findHeight int64
//		var findResult []byte
//		structBlock := block.Block{}
//		// Check the contract ID, if ID is not right return error
//		hasAdd := strings.Index(contractID, "+")
//		if hasAdd == -1 {
//			findResult = nil
//			return findResult, errors.New(" Contract ID is not right, please check it ")
//		}
//		// Get Hex Height from ID and transfer to Dec
//		splitContractID := bytes.Split([]byte(contractID), []byte("+"))
//		findHeight = utils.HexDec(string(splitContractID[1])) + 1
//		if findHeight == 0 {
//			findResult = nil
//			return findResult, errors.New(" Contract ID is not right, please check it ")
//		}
//		// Query contract loop
//		byteBlock, err := db_api.Db_getBlock(findHeight)
//		if err != nil {
//			findResult = nil
//			return findResult, err
//		}
//		for byteBlock != nil {
//			err = json.Unmarshal(byteBlock, &structBlock)
//			// Find the contract in this block
//			contract, ok := structBlock.TransactionMap[contractID]
//			// If the contract is in this block, just return
//			if ok {
//				findResult, _ = json.Marshal(contract)
//				break
//				// If the contract is not in this block, change block to next block
//			} else {
//				findHeight = findHeight + 1
//				byteBlock, err = db_api.Db_getBlock(findHeight)
//				if err != nil {
//					findResult = nil
//					return findResult, err
//				}
//			}
//		}
//		return findResult, nil
//	}(contractId)
//	if err != nil || jsonContractFound == nil {
//		errMsg := fmt.Sprintf(" %s", err.Error())
//		w.WriteHeader(http.StatusInternalServerError)
//		_, _ = w.Write([]byte("Contract ID : " + contractId + " query failed" + errMsg))
//		// If there is no error and contract is not nil
//	} else {
//		var contractFound block.Transaction
//		_ = json.Unmarshal(jsonContractFound, &contractFound)
//		contractQueryDataStruct := block.ContractQueryResponseData{
//			Contract: contractFound,
//		}
//		contractQueryResponseStruct := block.ContractQueryResponse{
//			Code: 0,
//			Data: contractQueryDataStruct,
//			Msg:  "...",
//		}
//		response, _ := json.MarshalIndent(contractQueryResponseStruct, "", "  ")
//		_, _ = w.Write(response)
//	}
//}
// 获取最新区块
func HandleGetNewBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	block.GChainMemRWlock.RLock()
	height := block.GChainMem.Height
	block.GChainMemRWlock.RUnlock()

	//read block info from local leveldb -- db
	blockJson, _ := db_api.Db_getBlock(height)
	blockTmp := block.Block{}
	_ = json.Unmarshal(blockJson, &blockTmp) //把json的字节流([]byte)解析到结构体里面

	type NewBlock struct {
		Height int64
		Block  block.Block
	}

	responseStruct := NewBlock{
		Height: height,
		Block:  blockTmp,
	}

	RespondWithJSON(w, r, http.StatusOK, responseStruct)

}
// 获得某个节点的本地区块信息
func HandleGetBlockInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	block.GChainMemRWlock.RLock()
	height := block.GChainMem.Height
	block.GChainMemRWlock.RUnlock()

	var count = 1
	var start int64
	var err error

	r.ParseForm()
	for k, v := range r.Form {
		switch k {
		case "count":
			count, err = strconv.Atoi(v[0])
			if err != nil {
				RespondWithJSON(w, r, http.StatusBadRequest, nil)
				return
			}
		}
	}

	start = height - int64(count) + 1
	if start < 1 {
		start = 1
	}

	utils.Log.Info("Http block height,", height)

	var allBlock []block.Block
	for i := start; i <= height; i++ {
		//read block info from local leveldb -- db
		blockJson, err := db_api.Db_getBlock(i)
		if err == nil {
			blockTmp := block.Block{}
			err = json.Unmarshal(blockJson, &blockTmp) //把json的字节流([]byte)解析到结构体里面
			if err == nil {
				blockTmp.Head.TxCount = len(blockTmp.TransactionMap)
				allBlock = append(allBlock, blockTmp)

			} else {
				utils.Log.Errorf("handleGetBlockInfo json unmarshal failed %s", err.Error())
			}
		}
	}

	allBlockChain := block.AllBlockInfo{
		Height: height,
		Block:  allBlock,
	}

	RespondWithJSON(w, r, http.StatusOK, allBlockChain)

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
		IsSubsidized bool
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
			Votes:        NodeAccount.Votes,
			Pledge:       NodeAccount.Pledge,
			IsSubsidized: NodeAccount.IsSubsidized,
		}
	}
	RespondWithJSON(w, r, http.StatusOK, NodeAccounts)

}

// 获取区块头信息
func HandleTestTps(w http.ResponseWriter, r *http.Request) {
	contract := test.TestTx{}
	contract.Timestamp1 = time.Now().UTC().UnixNano()
	contract.Timestamp2 = time.Now().UnixNano()
	contract.Timestamp3 = time.Now().Unix()
	bContract, _ := json.Marshal(contract)

	tx := createTx(bContract, block.TX_TEST)

	block.GTransactionPool.AddTransaction(*tx)
	block.P2pBroadcastTransaction(*tx,block.OpetatorAdd)

	RespondWithJSON(w,r,http.StatusOK,tx)

}

func HandleGetTxPool(w http.ResponseWriter, r *http.Request) {

	block.GTransactionPoolRWlock.RLock()

	txPool := block.GTransactionPool

	block.GTransactionPoolRWlock.RUnlock()

	RespondWithJSON(w,r,http.StatusExpectationFailed,txPool)
}

// DPOS 投票
//func HandleVoteTransaction(w http.ResponseWriter, r *http.Request) {
//
//	utils.Log.Info("Receive a Vote request")
//
//	w.Header().Set("Content-Type", "application/json")
//	var msg block.VoteRequest
//
//	decoder := json.NewDecoder(r.Body)
//	if err := decoder.Decode(&msg); err != nil {
//		RespondWithJSON(w, r, http.StatusBadRequest, r.Body)
//		utils.Log.Errorf("Vote request decode failed!! %s", err.Error())
//		return
//	}
//	defer r.Body.Close()
//
//	jsonCandidateAccount, err := db_api.Db_getAccount(string(msg.CandidateAddress))
//	if err != nil {
//		utils.Log.Errorf("handleGetBlockInfo read candidate from db failed %s", err.Error())
//		responseJson := block.PushTransactionResponse{
//			ResponseCode: "fail",
//			ExtraInfo:    "Bad candidate's address",
//		}
//		RespondWithJSON(w, r, http.StatusBadRequest, responseJson)
//		return
//	}
//	CandidateAccount := account.AccountInDb{}
//	err = json.Unmarshal(jsonCandidateAccount, &CandidateAccount)
//	if err != nil {
//		utils.Log.Errorf("handleGetBlockInfo json unmarshal failed %s", err.Error())
//		RespondWithJSON(w, r, http.StatusIMUsed, r.Body)
//		return
//	}
//
//	jsonVoterAccount, err := db_api.Db_getAccount(string(msg.VoterAddress))
//	if err != nil {
//		utils.Log.Errorf("handleGetBlockInfo read voter from db failed %s", err.Error())
//		responseJson := block.PushTransactionResponse{
//			ResponseCode: "fail",
//			ExtraInfo:    "Bad voter's address",
//		}
//		RespondWithJSON(w, r, http.StatusBadRequest, responseJson)
//		return
//	}
//	VoterAccount := account.AccountInDb{}
//	err = json.Unmarshal(jsonVoterAccount, &VoterAccount)
//	if err != nil {
//		utils.Log.Errorf("handleGetBlockInfo json unmarshal failed %s", err.Error())
//		RespondWithJSON(w, r, http.StatusIMUsed, r.Body)
//		return
//	}
//
//	if CandidateAccount.Balance >= float64(msg.Pledge) {
//
//		transaction := block.Transaction{}
//		//fixme need self-account transaction.SignedAccount
//		transaction.TimeStamp = time.Now().Unix()
//		transaction.TransactionID = "TIMESTAMP" + strconv.FormatInt(transaction.TimeStamp, 10) +
//			"VOTER_ADDR" + msg.VoterAddress
//		transaction.TransactionType = block.TransactionTypeVote
//		transaction.TransactionBody = block.VoteRequest{
//			CandidateAddress: msg.CandidateAddress,
//			VoterAddress:     msg.VoterAddress,
//			Pledge:           msg.Pledge,
//		}
//		//broadcast to others
//		Network.P2pBroadcastToSuperNodeTransaction(transaction, block.OpetatorAdd)
//
//		utils.Log.Info("Valid voting ! ", transaction)
//
//		//add to local pool
//		block.GTransactionPool.AddTransaction(transaction)
//
//		responseJson := block.PushTransactionResponse{
//			ResponseCode: "success",
//			ExtraInfo:    "None",
//		}
//		RespondWithJSON(w, r, http.StatusOK, responseJson)
//	} else {
//		utils.Log.Info("Balance is not enough")
//		responseJson := block.PushTransactionResponse{
//			ResponseCode: "fail",
//			ExtraInfo:    "Voter's balance is not enough",
//		}
//		RespondWithJSON(w, r, http.StatusBadRequest, responseJson)
//		//fix-me need to identify specific error
//	}
//
//}
