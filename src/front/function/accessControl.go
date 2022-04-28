package function

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/db_api"
	Network "DAG-Exp/src/network"
	"DAG-Exp/src/utils"
	"encoding/json"
	"net/http"
	"reflect"
	"time"
)

// 建立 DAG 头部 Tag
func HandleCreateDAG(w http.ResponseWriter, r *http.Request){
	/* 解析请求内容 */
	w.Header().Set("Content-Type", "application/json")
	rDecoder := json.NewDecoder(r.Body)
	var msg ReqCreateDAG
	err := rDecoder.Decode(&msg)
	if err != nil {
		resp := ResponseInfo{
			Message: "请求格式错误",
			Payload: nil,
		}
		RespondWithJSON(w, r, http.StatusBadRequest, resp)
		utils.Log.Errorf("Decode failed.[ %s ]", err)
		return
	}

	/* 转化成 TAG 数据结构 */
	tag := dag.DagTag{
		Depth: 		msg.Depth,
		PrevHash:  	msg.PrevHash,
		TagType:   	msg.TagType,
		Body:      	msg.Body,
		Miner:     	msg.Miner,
		Signature: 	msg.Signature,
	}
	hash := dag.CalculateTagHash(tag)
	tag.Hash = hash

	/* 发送给其他节点 */
	Network.BroadcastDagTag(tag)

	/* 存入数据库 */
	// Hash, Tag
	errSaveTag := db_api.LevelDbDagSaveTag(tag)

	if errSaveTag != nil {
		RespondWithJSON(w, r, http.StatusBadRequest, err)
	} else {
		RespondWithJSON(w, r, http.StatusAccepted, tag)
	}
	return
}

/* {
	"fromAddress":"1FCA8TaoQw7uRaKFdoJHB3eWbEAte5yzNq",
    "toAddress":"1CEKaGFuue6RSvS25BEwfnm6PDvDa1WDzN",
    "reqInfo":"Test",
    "token":"Test_Access_Token",
    "reqHash":""
}*/

func HandleRequestData(w http.ResponseWriter, r *http.Request){
	body := dag.RequestTagBody{
		FromAddress: "1FCA8TaoQw7uRaKFdoJHB3eWbEAte5yzNq",
		ToAddress:   "1CEKaGFuue6RSvS25BEwfnm6PDvDa1WDzN",
		ReqInfo:     "Test",
		Token:       "Test_Access_Token",
	}
	bBody, _ := json.Marshal(body)
	bHash, _ := utils.GetHashFromBytes(bBody)
	body.ReqHash = string(utils.Base58Encode(bHash))
	// 2.2 获取可连接 Tag Hash
	preTagHash, preDepth :=  db_api.GetAttachableTagRandomly()
	utils.Log.Debug("随即连接：", preTagHash)
	// 2.3 填充 Tag 数据结构
	tag := dag.DagTag{
		TimeStamp:  time.Now().UnixNano(),
		Depth:	 	preDepth + 1,
		PrevHash:  	preTagHash,
		TagType:   	dag.REQUESTTAG,
		Body:      	body,
		Miner:     	string(account.GAccount.Address),
		Signature: 	"Signature_" + string(account.GAccount.Address),
	}
	// 2.4 计算 Tag Hash 并填充
	tag.Hash = dag.CalculateTagHash(tag)
	utils.Log.Debug("Request Tag Hash：",tag.Hash)
	// 存入数据库
	_ = db_api.LevelDbDagSaveTag(tag)
	// 广播 Req Tag
	Network.BroadcastDagTag(tag)

	RespondWithJSON(w, r, http.StatusOK, tag)
	return
}


func HandleGetDag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// dagList := db_api.GetAllDag()

	type findDAGResult struct {
		N    int `json:"n"`
		NA   int `json:"na"`
		//List []dag.DagTag `json:"list"`
	}
	n := db_api.GetTagCount()
	keys := reflect.ValueOf(db_api.UnConnectedHashMap).MapKeys()

	f := findDAGResult{
		N:    n,
		NA:   len(keys),
		//List: dagList,
	}
	RespondWithJSON(w, r, http.StatusOK, f)
	return
}

func HandleLatency(w http.ResponseWriter, r *http.Request){
	/* 解析请求内容 */
	w.Header().Set("Content-Type", "application/json")

	var averageLatency float64

	tags,_ := db_api.LevelDbTagTableGetTypeList(dag.ACKTAG)
	m := make(map[string]int64)
	n := float64(len(tags))
	for _, v := range tags {
		m[v] = db_api.CalculateLatency(v)
		averageLatency = averageLatency + float64(m[v]) / n
	}

	type resp struct {
		AverageLatency 	float64				`json:"average_latency"`
		List 			map[string]int64	`json:"list"`
 	}

 	response := resp{
		AverageLatency: averageLatency,
		List:           m,
	}
	RespondWithJSON(w, r, http.StatusFound, response)
	return
}