package db_api

import (
	"bfc/src/utils"
	"encoding/json"
	"strconv"
	"sync"
	"time"
)
var LatencyLock sync.RWMutex
func SaveRequestTimeStamp(hash string, timestamp int64) {
	sTimestamp := strconv.FormatInt(timestamp, 10)
	_ = DbBlock.Put([]byte("Time"+hash), []byte(sTimestamp), nil)
}

func UpdateLatencyList(reqHash string, endT int64) {
	LatencyLock.Lock()
	// 计算对应时延
	var list []int64
	bList, _ := DbBlock.Get([]byte("Latency"), nil)
	_ = json.Unmarshal(bList, &list)
	l := GetLatency(reqHash, endT)

	// 添加到时延记录列表
	list = append(list, l)
	bList, _ = json.Marshal(list)
	_ = DbBlock.Put([]byte("Latency"), bList, nil)

	// 计算平均时延
	//var averageT float64
	//averageT = 0
	//for _, v := range list {
	//	averageT =  averageT + float64(v)/float64(len(list))
	//}
	//bAverageLatency := []byte(strconv.FormatFloat(averageT,'E',-1,64))
	//_ = DbBlock.Put([]byte("AverageLatency"), bAverageLatency, nil)

	LatencyLock.Unlock()

}

func GetLatencyList () []int64 {
	b, _ := DbBlock.Get([]byte("Latency"), nil)
	var list []int64
	_ = json.Unmarshal(b, &list)
	return  list
}

func GetLatency(reqHash string, endT int64) int64 {
	bStartT, _ := DbBlock.Get([]byte("Time"+reqHash), nil)
	startT, _ := strconv.ParseInt(string(bStartT),10,64)
	us := ( endT - startT ) / int64(time.Microsecond)
	utils.Log.Debug("交易：", reqHash,"开始时间：", startT,",结束时间：",endT, "时延：", us)
	return us
}

func GetAverageLatency() float64 {
	t, _ := DbBlock.Get([]byte("AverageLatency"), nil)
	latency, _ := strconv.ParseFloat(string(t),64)
	return latency
}