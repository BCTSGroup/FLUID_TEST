package db_api

import (
	"DAG-Exp/src/dag"
	"DAG-Exp/src/utils"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"math/rand"
	"strconv"
	"sync"
	"time"
)
var UNCONNECTEDLOCK sync.RWMutex
var AVAILABLETAGLOCK sync.RWMutex
var TAGCOUNT sync.RWMutex
var LevelDBAvailableTag *leveldb.DB
var GTAGCOUNT int

var UnConnectedHashMap 	map[string]int
var AvailableHash 	[]string
func LevelDBSaveAvailableTag(tag dag.DagTag)  {
	// 更新 Tag Count
	TAGCOUNT.Lock()
	GTAGCOUNT = GTAGCOUNT + 1
	TAGCOUNT.Unlock()

	if tag.TagType == dag.TOPTAG {
		AvailableHash = append(AvailableHash, tag.Hash)
		UnConnectedHashMap[tag.Hash] = tag.Depth
		return
	}

	payload := tag.Body.(dag.AckTagBody)
	utils.Log.Debug("当前 ACK.reqhash：", payload.ReqHash)
	req := GetTag(payload.ReqHash)
	utils.Log.Debug("当前 REQ：", req)

	keyL := req.PrevHash[0]
	keyR := keyL
	if len(req.PrevHash) == 2 {
		keyR = req.PrevHash[1]
	}
	// 更新可连接的列表
	AVAILABLETAGLOCK.Lock()
	AvailableHash = append(AvailableHash, tag.Hash)
	count := 0
	for i, v := range AvailableHash {
			if v == keyL || v == keyR {
				AvailableHash = append(AvailableHash[:i], AvailableHash[i+1:]...)
				count = count + 1
				if count == 2 {
					break
				}
			}
	}
	AVAILABLETAGLOCK.Unlock()
	// 更新未被连接的 TAG 列表
	UNCONNECTEDLOCK.Lock()
	UnConnectedHashMap[tag.Hash] = tag.Depth
	utils.Log.Debug("现有可被连接的TAG：", UnConnectedHashMap, AvailableHash)
	utils.Log.Debug("当前节点的左，右：", keyL, keyR)
	if _, ok := UnConnectedHashMap[keyL]; ok{
		//>> 存在键值 前序左交易
		delete(UnConnectedHashMap, keyL)
	}
	if _, ok := UnConnectedHashMap[keyR]; ok{
		//>> 存在键值 前序右交易
		delete(UnConnectedHashMap, keyR)
	}

	UNCONNECTEDLOCK.Unlock()

}


// 当前最大 Depth
func getAvailableTagDepth() int{
	bDepth, err := LevelDBAvailableTag.Get([]byte("Depth"), nil)
	if err != nil {
		utils.Log.Error("Cannot get the depth in LevelDB, Error: ", err)
		return -1
	}
	depth, _ := strconv.Atoi(string(bDepth))
	return depth
}

// 整个系统 Tag 的总数
func GetTagCount() int {
	return GTAGCOUNT
}

// 对应 Depth 下的所有 Tag Hash
func getTagListByDepth(depth int) []string{
	var hashList []string
	sDepth := strconv.Itoa(depth)
	bHashList, err := LevelDBAvailableTag.Get([]byte(sDepth), nil)
	if err != nil {
		utils.Log.Debug("Get hash list by depth failed. Error: ", err)
		return hashList
	}

	_ = json.Unmarshal(bHashList, &hashList)
	return hashList
}

func SaveUnconnectedTagMap(m map[string]int) {
	bMap, err := json.MarshalIndent(m,"","\t")
	utils.Log.Debug("更新未连接的TAG MAP：", m)
	err = LevelDBAvailableTag.Put([]byte("Unconnected"), bMap,nil)
	if err != nil {
		utils.Log.Error("Save Unconnected Tag Failed. Error: ", err)
	}
}
func GetUnconnectedTagMap() map[string]int{
	hashMap := make(map[string]int)
	bHashList, err := LevelDBAvailableTag.Get([]byte("Unconnected"), nil)
	if err != nil {
		utils.Log.Error(err.Error())
		return hashMap
	}
	_ = json.Unmarshal(bHashList, &hashMap)
	return hashMap
}
func GetAttachableTagRandomly() ([]string, int){
	var list []string
	list = make([]string, len(AvailableHash))
	_ = copy(list, AvailableHash)
	var hash []string
	if len(list) == 1 {
		hashL := list[0]
		hash = append(hash, hashL)
		depth := GetTag(hashL).Depth
		return hash, depth
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := len(list)

	nL := r.Intn(l)
	nR := r.Intn(l)
	for ; nL == nR; {
		nR = r.Intn(l)
	}

	hashL := list[nL]
	hashR := list[nR]

	depthL := GetTag(hashL).Depth
	depthR := GetTag(hashR).Depth
	depth := depthL
	if depthL < depthR {
		depth = depthR
	}

	hash = append(hash, hashL)
	hash = append(hash, hashR)

	return hash, depth
	//var hashL, hashR string
	//var hash []string
	//var depth int
	//var n int
	//if l < 6 {
	//	n = r.Intn(l)
	//} else {
	//	n = l - r.Intn(6)
	//}
	//hashL = AvailableHash[n]
	//if l == 1 {
	//	hash = append(hash, hashL)
	//	hashR := ""
	//	hash = append(hash, hashR)
	//	depth = GetTag(hashL).Depth
	//	utils.Log.Debug("只有一个可连接节点，hash:", hash,"depth: ", depth)
	//	return hash, depth
	//}
	//if l > 1 && n < l -1 {
	//	hashR = AvailableHash[n+1]
	//}
	//if l > 1 && n >= l - 1 {
	//	hashR = AvailableHash[n-1]
	//}
	//hash = append(hash, hashL)
	//hash = append(hash, hashR)
	//dl := GetTag(hashL).Depth
	//dr := GetTag(hashR).Depth
	//if dl > dr {
	//	depth = dl
	//} else {
	//	depth = dr
	//}
}