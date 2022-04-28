package test

import (
	http2 "bfcTpsTest/http"
	"bfcTpsTest/parameter"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)


var LastHeight 		int64
var CurrentHeight   int64
// TODO : 创建一个结果数组，把所有TPS测试结果存入数组中并存储到文件中
// 吞吐量测试
func TimerRun () {
	fileName := "/Users/nijunpei/go/src/bfcTpsTest/SingleChain_TPS_" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	f,err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		panic("File created wrong." + err.Error())
	}
	writer := bufio.NewWriter(f)
	// Set Timer => 3s
	TimerDemo := time.NewTimer(time.Duration(1) * time.Second)
	// Timer loop
	for {
		select {
		case <-TimerDemo.C: {
			currentTps := countTps()
			if currentTps != 0 {
				tpsSave := strconv.FormatFloat(currentTps,'f',-1,64) + "\n"
				writer.WriteString(tpsSave)
				writer.Flush()
				fmt.Println("Tps : ", currentTps)
			}

			TimerDemo.Reset(time.Duration(3) * time.Second)
		}
		}
	}
}
func countTps () float64 {
	// Get the highest block in the chain
	getBlockInfo := getNewBlock()
	if getBlockInfo == nil {
		panic("[ ERROR : Cannot get the block info from bfc] === > Please reset bfc and the test.")
	}
	CurrentHeight = getBlockInfo.Height
	// If there is no new block, tps = 0
	if CurrentHeight == LastHeight {
		return 0
	}
	// If there is a new block, tps = txCount / 3sec
	txCount := len(getBlockInfo.Block.TransactionMap)
	tps := float64(txCount) / 3
	// Set LastHeight = CurrentHeight
	LastHeight = CurrentHeight
	// Return tps result
	return tps
}
func getNewBlock  () *parameter.GetNewBlockResp {
	// Get the highest block in the chain
	var getNewBlockResp parameter.GetNewBlockResp
	response := http2.Get("http://127.0.0.1:8000/GetNewBlock")
	if response == nil {
		return nil
	}
	_ = json.Unmarshal(response, &getNewBlockResp)
	return &getNewBlockResp
}
