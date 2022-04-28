package block

import (
	"bfc/src/db_api"
	"bfc/src/sharecc"
	"bfc/src/utils"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// 该合约考虑到经过网络传输的 interface{} 会变成 map[string]interface{}
// 设置共享规则合约，在接入请求被确认后，执行该合约
func Contract_SetSharingRule (from string, rule sharecc.JoinPayload) {
	// 设置系统规则
	err := db_api.SysCfgManager.SetShareRule(from, rule)
	if err != nil {
		utils.Log.Error(err)
	}
	return
}
/*	接收数据提交合约
	调用时机：	1. Http 接口收到提交请求
			 	2. 节点收到提交请求信息的广播

	调用结果：	1. 将请求中的数据记录调用外部验证合约进行验证
				2. 将通过验证的请求打包成交易
				3. 将共享数据临时存储在数据库中
				4. 将共享数据在数据库中的存储生成验证信息索引
				5. 将交易添加到交易池
*/
func Contract_ReceiveSubmitData (request sharecc.SubmitRequest) bool {
	// 1. todo 调用外部验证
	isVerified := sharecc.DataValidation(request.From, request)
	if !isVerified {return false}

	for _, v := range request.Payload.Data {
		// 2. 构造 Transaction，每一条数据构成一个Tx
		txSharingDataBody := &sharecc.TxSharingData{
			DataType:  v.DataType,
			Key:       v.Key,
			Value:     v.Value,
			TimeValid: v.TimeValid,
		}

		bTxSharingDataBody, _ := json.Marshal(txSharingDataBody)
		bTxHash, _ := utils.GetHashFromBytes(bTxSharingDataBody)
		txHash := hex.EncodeToString(bTxHash)
		utils.Log.Critical("收数据打包交易时 Hash，数据: ", txHash, string(bTxSharingDataBody))
		tx := &Transaction{
			TransactionID:   txHash,
			TimeStamp:       time.Now().UTC().UnixNano(),
			TransactionType: TX_SUBMIT,
			TransactionBody: txSharingDataBody,
		}

		// 3. 把记录临时存储在数据库中 & 4. 数据库加一区块索引记录
		err := db_api.ShareManager.TempSaving(request.From, *txSharingDataBody)
		if err != nil{
			msg,_ := json.MarshalIndent(txSharingDataBody,"","   ")
			utils.Log.Error("共享数据存储临时失败"+string(msg))
			continue
		}

		// 5. 加入到交易池
		GTransactionPool.AddTransaction(*tx)
	}
	return true
}
/*	共享数据存储合约
	调用时机：数据提交合约执行完，Tx通过共识后触发
	调用结果：数据存储等级 X -> Y
*/
func Contract_SaveSharingData(blockHash string, data sharecc.TxSharingData) {
	err := db_api.ShareManager.SaveSharingData(blockHash, data)
	if err != nil {
		utils.Log.Error("存储合约执行失败", err)
	}
}
/*	数据共享合约
	调用时机：	Http 接口接收到数据共享请求后
	调用结果：	1. 调用 db_api 找到相关请求的数据记录
				2. 调用 db_api 找到所属元数据所属区块，提取 Merkle 路径和Block hash
				3. 组装上述信息内容并返回
	返回信息：	1. 数据共享请求的返回Payload
*/
func Contract_DataSharing(getRequest sharecc.GetSharedDataRequest) *sharecc.GetResponsePayload {
	// 1. 根据key值从数据库中检索出相应数据记录
	data := db_api.ShareManager.GetSharedData(getRequest.Payload.Key)
	if data == nil {
		utils.Log.Error("没有检索到相关数据")
		return nil
	}
	// 如果是临时存储，则不得访问
	if strings.Index(data.Level, "Y") == -1{
		utils.Log.Error("数据暂时不可用")
		return nil
	}
	// 2. 提取验证信息
	payload := &sharecc.GetResponsePayload{
		DataType:   data.DataType,
		Key:        getRequest.Payload.Key,
		Value:      data.Value,
		TimeValid:  data.ValidTime,
		BlockHash:  data.BlockHash,
	}
	// Todo 根据 Block Hash 和数据建立 Merkle path
	txSharedData := sharecc.TxSharingData{
		DataType:  data.DataType,
		Key:       getRequest.Payload.Key,
		Value:     data.Value,
		TimeValid: data.ValidTime,
	}
	proofSet, proofIndex, numLeaves :=GetProofBySharedData(txSharedData, data.BlockHash)
	// 3. 组装消息 -> Response payload struct
	payload.MerklePath.ProofSet = proofSet
	payload.MerklePath.ProofIndex = proofIndex
	payload.MerklePath.NumLeaves = numLeaves
	return payload
}
/* 	共享记录合约
	调用时机：	1. 前端接收共享请求时，需要被记录时
				2. 其他节点收到类型为共享记录的 Message 时
	调用结果：	将共享记录打包成交易加入到交易池
*/
func Contract_DataSharingRecord(getRequest sharecc.GetSharedDataRequest) {

	data := db_api.ShareManager.GetSharedData(getRequest.Payload.Key)
	if data == nil {
		utils.Log.Error("没有找到共享数据记录")
		return
	}
	level := data.Level

	if strings.Index(level, "Y") == -1 {
		utils.Log.Error("执行数据共享记录合约没有找到相关的等级信息，或数据不能被共享")
		return
	}
	switch level {
	case "Y00":
		// 如果是 Y00 则不需要被记录
		return
	case "Y01":
		// 如果是 Y01 则需要被记录共享行为
		// 1. 构建共享记录 Tx
		txBody := &sharecc.DataSharingRecord{
			Requester:  getRequest.From,
			RequestKey: getRequest.Payload.Key,
			Timestamp:  getRequest.TimeStamp,
		}
		bTxBody, _ := json.Marshal(txBody)
		bTxHash, _ := utils.GetHashFromBytes(bTxBody)
		txHash := hex.EncodeToString(bTxHash)

		tx := Transaction{
			TransactionID:   txHash,
			TimeStamp:       time.Now().UTC().UnixNano(),
			TransactionType: TX_SHARE,
			TransactionBody: txBody,
		}

		GTransactionPool.AddTransaction(tx)
		// 2. 加入交易池
	case "Y10":
		// TODO 暂时不考虑身份系统
	case "Y11":
		// TODO 暂时不考虑身份系统

	}
}