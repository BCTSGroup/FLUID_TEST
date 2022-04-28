package function

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/db_api"
	Network "DAG-Exp/src/network"
	"net/http"
	"time"
)

func HandleSubmit(w http.ResponseWriter, r *http.Request){
	epochTag := dag.EpochTag{
		TimeStamp: 	  time.Now().UnixNano(),
		Hash:         "",
		PrevHashList: nil,
		Checksum:     "",
		Signature:    "",
	}
	//max := 0
	//for k, v := range db_api.UnConnectedHashMap {
	//	epochTag.PrevHashList = append(epochTag.PrevHashList, k)
	//	if v > max {
	//		max = v
	//	}
	//}
	epochTag.Signature = string(account.GAccount.Address)
	epochTag.Hash = dag.CalculateEpochTagHash(epochTag)
	dag.VOTELOCK.Lock()
	dag.VOTE[epochTag.Hash] = dag.VOTE[epochTag.Hash] + 1
	dag.VOTELOCK.Unlock()
	Network.BroadcastEpoch(epochTag)

	RespondWithJSON(w, r, http.StatusOK, epochTag)
	return
}

func HandleEpochLatency(w http.ResponseWriter, r *http.Request){
	type l struct {
		// AverageLatency map[string]float64
		Latency []float64
		Latency2 []int64
		//LatencyMap map[string][]int64
	}
	resp := l{}
	dag.VSLOCK.Lock()
	lm := dag.VoteLatencyResult

	var average float64
	var latency int64
	for _, v := range lm{
		latency = v[0]
		for _, value := range v {
			average = average + float64(value) / float64(len(v))
			if value < latency {
				latency = value
			}
		}
		// resp.AverageLatency[k] = average
		resp.Latency = append(resp.Latency, average)
		resp.Latency2 = append(resp.Latency2, latency)
		average = 0
	}
	dag.VSLOCK.Unlock()
	//resp.LatencyMap = lm
	RespondWithJSON(w, r, http.StatusOK, resp)
	return
}


func HandleGetTips(w http.ResponseWriter, r *http.Request){
	RespondWithJSON(w, r, http.StatusOK, db_api.UnConnectedHashMap)
	return

}