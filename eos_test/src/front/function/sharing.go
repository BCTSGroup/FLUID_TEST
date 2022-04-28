package function

import (
	"bfc/src/block"
	"bfc/src/db_api"
	"bfc/src/test"
	"bfc/src/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"

	"bfc/src/sharecc"
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

func HandleJoinRequest(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	// 解析 Request Body
	rDecoder := json.NewDecoder(r.Body)
	var joinRequest sharecc.JoinRequest
	err := rDecoder.Decode(&joinRequest)
	if err != nil {
		RespondWithJSON(w, r, http.StatusBadRequest, r.Body)
		utils.Log.Errorf("Join request decode failed.[ %s ]", err)
		return
	}
	// TODO 解析数据并第一次调用外部验证模块 (伪代码 为了流程逻辑的完整)
	ok := sharecc.DataValidation(joinRequest.From, joinRequest)
	if !ok {
		failResp := response{
			Status: FAIL,
			Prompt: "Create transaction failed. Please try again.",
		}
		RespondWithJSON(w, r, http.StatusBadRequest, failResp)
		return
	}
	// 填写 TX 表单
	bContract, _ := json.Marshal(joinRequest)
	txJoin := createTx(bContract, block.TX_JOIN)
	if txJoin == nil {
		// 消息处理失败，反馈重新提交信息
		failResp := response{
			Status: FAIL,
			Prompt: "Create transaction failed. Please try again.",
		}
		RespondWithJSON(w, r, http.StatusBadRequest, failResp)
		return
	}
	// 加入本地交易池
	block.GTransactionPool.AddTransaction(*txJoin)
	// 交易消息广播给其他节点
	block.P2pBroadcastTransaction(*txJoin, block.OpetatorAdd)
	// 返回请求提交成功反馈信息
	successResp := response{
		Status: SUCCESS,
		Prompt: "Request submitted successfully.",
	}
	RespondWithJSON(w, r, http.StatusOK, successResp)
	return
}

func HandleSearchRule(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	from := vars["From"]
	rule := db_api.SysCfgManager.GetRule(from)
	if rule == nil {
		responseMsg := response{
			Status: FAIL,
			Prompt: "没有找到该链的共享规则",
		}
		RespondWithJSON(w, r, http.StatusNotFound, responseMsg)
		return
	}
	responseMsg := response{
		Status: SUCCESS,
		Prompt: rule,
	}
	RespondWithJSON(w, r, http.StatusOK, responseMsg)
}

func HandleSubmitRequest(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	// 解析 Submit Request
	rDecoder := json.NewDecoder(r.Body)
	var submitRequest sharecc.SubmitRequest
	err := rDecoder.Decode(&submitRequest)
	if err != nil {
		RespondWithJSON(w, r, http.StatusBadRequest, r.Body)
		utils.Log.Error("Submit request 解析失败")
		return
	}
	// 验证消息来源
	isVerified := sharecc.DataValidation(submitRequest.From, submitRequest.Payload)
	if !isVerified {}
	// 本地执行提交请求合约
	isExecd := block.Contract_ReceiveSubmitData(submitRequest)
	if !isExecd {
		failResponse := response{
			Status: FAIL,
			Prompt: "处理数据提交合约执行失败!",
		}
		RespondWithJSON(w, r, http.StatusBadRequest, failResponse)
		return
	}
	// 广播请求原始消息
	block.P2pBroadcastSubmitSharingDataRequest(submitRequest)

	successResponse := response{
		Status: SUCCESS,
		Prompt: "数据提交请求成功",
	}
	RespondWithJSON(w, r, http.StatusOK, successResponse)
}

func HandleGetSharingData(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	// 解析 Submit Request
	rDecoder := json.NewDecoder(r.Body)
	var getDataRequest sharecc.GetSharedDataRequest
	err := rDecoder.Decode(&getDataRequest)
	if err != nil {
		RespondWithJSON(w, r, http.StatusBadRequest, r.Body)
		utils.Log.Error("Get shared data request 解析失败")
		return
	}
	// 广播数据请求
	block.P2pBroadcastGetSharedDataRequest(getDataRequest)

	// 调用共享数据合约
	payload := block.Contract_DataSharing(getDataRequest)
	if payload == nil {
		failResponse := response{
			Status: FAIL,
			Prompt: "没有相关数据!",
		}
		RespondWithJSON(w, r, http.StatusBadRequest, failResponse)
		return
	}
	getDataResponse := sharecc.GetResponse{
		Code:      http.StatusOK,
		Data:      *payload,
		Timestamp: time.Now().UTC().UnixNano(),
	}

	// 返回共享数据合约查找结果
	RespondWithJSON(w, r, http.StatusOK, getDataResponse)

	// 调用记录合约
	block.Contract_DataSharingRecord(getDataRequest)

	return
}

// For test 同步区块头请求
func HandleSyncBlockHead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	localHeight := vars["LocalHeight"]
	height, err := strconv.ParseInt(localHeight, 10, 64)
	if err != nil {
		utils.Log.Error("高度转换数字类型失败", err)
		responseMsg := response{
			Status: FAIL,
			Prompt: "本地区块高度解析错误",
		}
		RespondWithJSON(w, r, http.StatusExpectationFailed, responseMsg)
		return
	}
	// 计算其实高度和获取最新区块高度作为终止高度
	start := height + 1
	block.GChainMemRWlock.RLock()
	end := block.GChainMem.Height
	block.GChainMemRWlock.RUnlock()

	// For test 定义返回结构体
	type BlockHeadInfo struct {
		// HashToHeight map[Hash]Height
		EndHeight	  int64	`json:"end_height"`
		HashToHeight  map[string]string				`json:"hash_to_height"`
		BlockHeadList map[string]block.BlockHead	`json:"block_head_list"`
	}
	hash2Height := make(map[string]string)
	blockList := make(map[string]block.BlockHead)
	// 循环获取 Block Head
	for i := start; i <= end; i++ {
		// 1. 从数据库中获取高度为 i 的区块
		blockJson, err := db_api.Db_getBlock(i)
		if err != nil {
			utils.Log.Error("由区块高度获取对应区块失败", err)
			responseMsg := response{
				Status: FAIL,
				Prompt: "获取区块信息失败",
			}
			RespondWithJSON(w, r, http.StatusExpectationFailed, responseMsg)
			break
		}
		// 2. 解析区块信息到 block struct
		var localBlock block.Block
		json.Unmarshal(blockJson, &localBlock)
		// 3. 解析区块信息到定义的结构体内容
		h := strconv.FormatInt(i, 10)
		blockList[h] = localBlock.Head
		hash2Height[localBlock.Head.Hash] = h
	}
	respBlockHeadInfo := BlockHeadInfo{
		EndHeight:	   end,
		HashToHeight:  hash2Height,
		BlockHeadList: blockList,
	}

	RespondWithJSON(w, r, http.StatusOK, respBlockHeadInfo)
}

func HandleRequestData(w http.ResponseWriter, r *http.Request) {
	req := test.TXRequestData{
		Timestamp:   time.Now().UTC().UnixNano(),
		FromAddress: "1FCA8TaoQw7uRaKFdoJHB3eWbEAte5yzNq",
		ToAddress:   "1CEKaGFuue6RSvS25BEwfnm6PDvDa1WDzN",
		ReqInfo:     "Test_Req",
		Token:       "ABC-ABC",
	}
	bReq, _ := json.Marshal(req)
	bHash, _ := utils.GetHashFromBytes(bReq)
 	req.ReqHash= string(utils.Base58Encode(bHash))

 	// 组建 Transaction
 	bTxBody,_ := json.MarshalIndent(req,"","\t")
 	tx := createTx(bTxBody, block.TX_REQUEST)
	block.GTransactionPool.AddTransaction(*tx)

	block.P2pBroadcastTransaction(*tx, block.OpetatorAdd)
	successResp := response{
		Status: SUCCESS,
		Prompt: "Request submitted successfully.",
	}
	utils.Log.Debug("交易：", req.ReqHash, "请求时间：", req.Timestamp)
	RespondWithJSON(w, r, http.StatusOK, successResp)
	return
}

func HandleGetLatency(w http.ResponseWriter, r *http.Request) {
	list := db_api.GetLatencyList()
	type resp struct {
		AverageLatency float64 `json:"average_latency"`
		LatencyList []int64 	`json:"latency_list"`
	}
	avg := float64(0)
	for _, v := range list {
		avg = avg + float64(v)
	}
	avg = avg / float64(len(list))
	avg = avg / 1000
	responseMsg := resp{
		AverageLatency: avg,
		LatencyList:    list,
	}
	RespondWithJSON(w,r,http.StatusOK,responseMsg)
}