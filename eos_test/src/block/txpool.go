package block

import (
	"bfc/src/db_api"
	"bfc/src/test"
	"bfc/src/utils"
	"encoding/hex"
	"encoding/json"
	"time"
)

// 交易池结构体
type TransactionPool struct {
	TransactionNum int64
	TransactionMap map[string]Transaction // (k, v) = (txID, Tx)
}


// 将交易添加到交易池
func (Txpool *TransactionPool) AddTransaction(transaction Transaction) {
	// 解析交易类型
	// TODO 如果是 Submit data 在这里重构交易体 ？
	// 加入交易池 Map
	GTransactionPoolRWlock.Lock()

	// bTx, _ := json.MarshalIndent(transaction,"","   ")
	// utils.Log.Debug("添加交易到交易池",string(bTx))

	_, ok := GTransactionPool.TransactionMap[transaction.TransactionID]
	if !ok {
		GTransactionPool.TransactionNum += 1
	}
	GTransactionPool.TransactionMap[transaction.TransactionID] = transaction

	GTransactionPoolRWlock.Unlock()
}

//当把交易加入区块中不是将交易从交易池移除的时机，只有区块变成不可逆的时候，执行交易并把它从交易池中移除
func (Txpool *TransactionPool) CommitTransactionAndExecuteInBlock(transactionListNowBlock map[string]Transaction, blockHash string) {
	//i := 0
	GTransactionPoolRWlock.RLock()
	utils.Log.Critical("交易池中内含交易数：", GTransactionPool.TransactionNum)
	//for k := range GTransactionPool.TransactionMap {
	//	utils.Log.Infof(" %d  Commit Transaction And Execute In Block: %s", i, k)
	//	i++
	//}
	GTransactionPoolRWlock.RUnlock()

	for _, transactionI := range transactionListNowBlock {

		// 遍历区块中的Transaction，处理并从交易池中删除
		// 遍历交易池,寻找相同的,用的map,所以时间复杂度是O(N)
		GTransactionPoolRWlock.Lock()
		// utils.Log.Info(GTransactionPool.TransactionNum,GTransactionPool.TransactionMap)
		if _, ok := GTransactionPool.TransactionMap[transactionI.TransactionID]; ok {
			// 从交易池中删除
			delete(GTransactionPool.TransactionMap, transactionI.TransactionID)
			GTransactionPool.TransactionNum -= 1
		}
		GTransactionPoolRWlock.Unlock()

		// 判断类型并本地执行交易内容
		switch transactionI.TransactionType {
		// TODO 加入测试逻辑
		//case TX_JOIN:{
		//	// interface{} 读取 from
		//	txBodyMap := transactionI.TransactionBody.(map[string]interface{})
		//	from := txBodyMap["from"].(string)
		//	// interface{} 转 JoinRequest
		//	txBodyPayloadMap := txBodyMap["payload"]
		//	btxBodyPayload, _ := json.Marshal(txBodyPayloadMap)
		//	var txBodyPayload sharecc.JoinPayload
		//	json.Unmarshal(btxBodyPayload, &txBodyPayload)
		//	// 执行共享规则设置合约
		//	Contract_SetSharingRule(from, txBodyPayload)
		//}
		//case TX_SUBMIT:{
		//	// 执行共享数据提交存储合约
		//	// interface{} 转 TxSharingData
		//	bTxSharedData, _ :=json.Marshal(transactionI.TransactionBody)
		//	txSharedData := sharecc.TxSharingData{}
		//	json.Unmarshal(bTxSharedData, &txSharedData)
		//
		//	utils.Log.Critical(string(bTxSharedData))
		//
		//	// 调用存储合约
		//	Contract_SaveSharingData(blockHash, txSharedData)
		//}
		case TX_REQUEST:{
			txReqBody := transactionI.TransactionBody.(map[string]interface{})


			toAddress := txReqBody["toAddress"].(string)
			fromAddress := txReqBody["fromAddress"].(string)
			reqHash := txReqBody["reqHash"].(string)
			startT := int64(txReqBody["timestamp"].(float64))
			db_api.SaveRequestTimeStamp(reqHash, startT)
			utils.Log.Debug("存储请求数据开始时间：(hash, T)",reqHash,",",startT)
			if GMyAccountAddress == toAddress {
				txResponseBody := test.TXResponseData{
					Timestamp:   time.Now().UTC().UnixNano(),
					FromAddress: GMyAccountAddress,
					ToAddress:   fromAddress,
					ExecResult:  false,
					InfoURL:     "xxx",
					Code:        "xxx",
					ReqHash:     reqHash,
				}
				bRespBody, _ := json.Marshal(txResponseBody)
				bTxHash, _ := utils.GetHashFromBytes(bRespBody)
				txHash := hex.EncodeToString(bTxHash)
				tx := Transaction{
					TransactionID:   txHash,
					TimeStamp:       time.Now().UTC().UnixNano(),
					TransactionType: TX_RESPONSE,
					TransactionBody: txResponseBody,
				}
				GTransactionPool.AddTransaction(tx)
				P2pBroadcastTransaction(tx, OpetatorAdd)
			}
		}
		case TX_RESPONSE:{
			txResponseBody := transactionI.TransactionBody.(map[string]interface{})
			endT := time.Now().UTC().UnixNano()
			utils.Log.Debug("交易：", txResponseBody["reqHash"].(string), "完成时间：", endT)
			db_api.UpdateLatencyList(txResponseBody["reqHash"].(string), endT)
		}
		default:
		}
	}
	CONSENSUS_MAP_LOCK.Lock()
	CONSENSUS_MAP[blockHash] = true
	CONSENSUS_MAP_LOCK.Unlock()

	GTransactionPoolRWlock.RLock()
	utils.Log.Critical("交易执行完之后交易池内含交易数：", GTransactionPool.TransactionNum)
	GTransactionPoolRWlock.RUnlock()


}
