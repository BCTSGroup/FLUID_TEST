package main

import (
	"bfc/src/utils"
	"encoding/json"
	"io/ioutil"
)

func main() {
	dir, err := ioutil.ReadDir("../node")
	if err != nil {
		utils.Log.Fatalf("读取文件夹目录出错")
		return
	}
	//PthSep := string(os.PathSeparator)
	//util.Log("[",utils.GetUTCTime(),"]  ","[",utils.GetUTCTime(),"]  ",PthSep)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	var nodeMap map[string]string
	nodeMap = make(map[string]string)
	for _, fi := range dir {
		var targetFileName string
		var folderName string
		if fi.IsDir() { // 目录, 递归遍历
			folderName = fi.Name()
			targetFileName = "../node/" + folderName + "/log_compare_temp.log"
			gotBlockInfo, e := ioutil.ReadFile(targetFileName)
			if e != nil {
				utils.Log.Fatalf("读取文件出错")
				return
			}

			nodeMap[folderName] = string(gotBlockInfo)

		}
	}

	//var node_map_json map[string] map

	//node是节点名字，test1，test2等
	//下面把字符串转成json
	var nodeJsonMap map[string]interface{}
	nodeJsonMap = make(map[string]interface{})
	for node := range nodeMap {

		var (
			nodeJson map[string]interface{}
		)
		//nodeJson=make(map[string]interface{})
		if err := json.Unmarshal([]byte(nodeMap[node]), &nodeJson); err == nil {
			nodeJsonMap[node] = nodeJson
			utils.Log.Infof("%v", nodeJson)
		}
	}

	utils.Log.Info(nodeJsonMap["test1"])
	//进行对比所有的节点数据是否一致
	//lowestHeight:=0
	//for node := range nodeJsonMap {
	//	tmp_JSON_map:=nodeJsonMap[node]
	//	//if tmp_JSON_map["Height"]>lowestHeight {
	//	//
	//	//}
	//util.Log(("[",utils.GetUTCTime(),"]  ",tmp_JSON_map)
	//}

}
