package test

import (
	http2 "bfcTpsTest/http"
	"encoding/json"
	"fmt"
)
type proposalLatency struct {
	// AverageLatency map[string]float64
	Latency []float64
	Latency2 []int64
	//LatencyMap map[string][]int64
}
func ProposeTxLatency() {
	var p proposalLatency
	response := http2.Get("http://127.0.0.1:7999/EpochLatency")
	_ = json.Unmarshal(response, &p)
	t := float64(0)
	for _, v := range p.Latency {
		t = t + v
	}
	t = t / float64(len(p.Latency))
	fmt.Println(t)
	t = float64(0)
	for _, v := range p.Latency2 {
		t = t + float64(v)
	}
	t = t / float64(len(p.Latency))
	fmt.Println(t)
}