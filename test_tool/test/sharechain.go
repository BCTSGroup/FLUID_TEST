package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const MULTI = 0
const SHC   = 1
//func init () {
//	// Init the test, set LastHeight as the height of BlockChain right now.
//	for {
//		firstGetBlock := getNewBlock()
//		if firstGetBlock != nil {
//			LastHeight = firstGetBlock.Height
//			break
//		} else {
//			fmt.Println("服务端未启动")
//			os.Exit(0)
//		}
//	}
//}
// 请求共享数据的返回结果
type GetResponse struct {
	Code		int					`json:"code"`
	Data		GetResponsePayload	`json:"data"`
	Timestamp	int64				`json:"timestamp"`
}
type GetResponsePayload struct {
	DataType    string 	`json:"data_type"`
	Key		  	string	`json:"key"`
	Value 	  	string	`json:"value"`
	TimeValid   string  `json:"time_valid"`
	BlockHash	string	`json:"block_hash"`
	MerklePath	MerklePath	`json:"merkle_path"`
}
type MerklePath struct {
	// ProofSet 就是验证用的路径 其中ProofSet[0] 就是需要验证的数据的 []byte 形式
	ProofSet  	[][]byte	`json:"proof_set"`
	ProofIndex 	uint64		`json:"proof_index"`
	NumLeaves	uint64		`json:"num_leaves"`
}
// 请求共享数据结构体
type GetSharedDataRequest struct {
	MsgType 	int 					`json:"msg_type"`
	From    	string  				`json:"from"`
	Payload 	GetSharedDataPayload 	`json:"payload"`
	TimeStamp 	int64					`json:"time_stamp"`
}
type GetSharedDataPayload struct {
	DataType	string	`json:"data_type"`
	Key			string	`json:"key"`
	Signature 	string	`json:"signature"`
}
const (
	JOIN = iota
	SUBMIT
	GET_BLOCKHEAD
	GET_DATA
)

func TestLatency(n int, strategy uint) {
	switch strategy {
	case SHC:
		t1 := time.Now() // get current time

		//logic handlers
		for i := 0; i < n; i++ {
			GetSharedData("http://127.0.0.1:7999/GetSharingData", "Car_AE86", "Address-Client-A")
		}
		elapsed := time.Since(t1)
		fmt.Println("App elapsed: ", elapsed)
	case MULTI:
		t := time.Now()
		t1 := time.Now() // get current time
		//logic handlers
		for i := 0; i < n/3; i++ {
			GetSharedData("http://127.0.0.1:7999/GetSharingData", "Car_AE86", "Address-Client-A")
		}
		elapsed := time.Since(t1)
		fmt.Println("第1次查询时间: ", elapsed)

		t2 := time.Now() // get current time
		//logic handlers
		for i := 0; i < n/3; i++ {
			GetSharedData("http://127.0.0.1:7999/GetSharingData", "Car_AE86", "Address-Client-A")
		}
		elapsed = time.Since(t2)
		fmt.Println("第2次查询时间: ", elapsed)

		t3 := time.Now() // get current time
		//logic handlers
		for i := 0; i < n/3; i++ {
			GetSharedData("http://127.0.0.1:7999/GetSharingData", "Car_AE86", "Address-Client-A")
		}
		elapsed = time.Since(t3)
		fmt.Println("第3次查询时间: ", elapsed)

		elapsed = time.Since(t)
		fmt.Println("总查询时间：",elapsed)

	}

}

// 请求共享数据
func GetSharedData(url string, key string, address string) *GetResponsePayload {
	// 1. 组装请求结构体
	// todo 概念验证实验，省略签名
	requestBody := GetSharedDataRequest{
		MsgType:   GET_DATA,
		From:      address,
		TimeStamp: time.Now().UTC().UnixNano(),
	}
	payload := GetSharedDataPayload{
		DataType:  "",
		Key:       key,
		Signature: "",
	}
	requestBody.Payload = payload

	// 2. 序列化请求结构体
	bRequestBody, _ := json.Marshal(requestBody)

	// 3. 建立客户端
	client := &http.Client{}
	req := bytes.NewBuffer(bRequestBody)
	request, _ := http.NewRequest("POST", url, req)
	request.Header.Set("Content-type", "application/json")

	// 4. 执行请求
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("请求出错", err)
		return nil
	}
	respBody, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(string(respBody))
	var getResponse GetResponse
	err = json.Unmarshal(respBody, &getResponse)
	if err != nil {
		fmt.Println(err)
	}
	return &getResponse.Data
}