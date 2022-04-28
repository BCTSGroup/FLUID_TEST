package test

import (
	http2 "bfcTpsTest/http"
	"bfcTpsTest/parameter"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var CurrentTagCount int
var LastTagCount int
// TODO : 创建一个结果数组，把所有TPS测试结果存入数组中并存储到文件中

func DAGTimerRun () {
	fileName := "/Users/nijunpei/go/src/bfcTpsTest/DAG_TPS_" + "5000"
	f,err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		panic("File created wrong." + err.Error())
	}
	//writer := bufio.NewWriter(f)
	// Set Timer => 3s
	TimerDemo := time.NewTimer(time.Duration(500) * time.Millisecond)
	// Timer loop
	for {
		select {
		case <-TimerDemo.C: {
			// currentTps := countDAGTps()
			//if currentTps != 0{
			//	fmt.Println("Tps : ", currentTps)
			//	//tpsSave := strconv.FormatFloat(currentTps,'f',-1,64) + "\n"
			//	//writer.WriteString(tpsSave)
			//	//writer.Flush()
			//}

			currentTips := countTips()
			fmt.Println(currentTips)

			TimerDemo.Reset(time.Duration(500) * time.Millisecond)
		}
		}
	}
}
func countTips() int {
	var tagCount parameter.DAGTagCount
	response := http2.Get("http://127.0.0.1:8000/GetDag")

	if response == nil {
		return 0
	}
	_ = json.Unmarshal(response, &tagCount)

	return tagCount.NA
}

func countDAGTps () float64 {

	CurrentTagCount = getDAGTAGCount()
	if CurrentTagCount == -1 {
		panic("[ ERROR : Cannot get the block info from bfc] === > Please reset bfc and the test.")
	}


	tagCount := CurrentTagCount - LastTagCount
	tps := float64(tagCount) // "/ 3"
	//
	LastTagCount = CurrentTagCount
	// Return tps result
	return tps
}
func  getDAGTAGCount  () int {
	// Get the highest block in the chain
	var tagCount parameter.DAGTagCount
	response := http2.Get("http://127.0.0.1:8000/GetDag")

	if response == nil {
		return 0
	}
	_ = json.Unmarshal(response, &tagCount)

	return tagCount.N

}