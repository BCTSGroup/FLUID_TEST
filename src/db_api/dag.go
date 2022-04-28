package db_api

import (
	"DAG-Exp/src/dag"
	"DAG-Exp/src/utils"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
)

var LevelDbDagInfo *leveldb.DB

func init(){
	UnConnectedHashMap = make(map[string]int)
}
func LevelDbDagSaveTag(tag dag.DagTag) error {
	// 直接存储 Tag
	hash := tag.Hash
	jsonTag, err := json.MarshalIndent(tag,"","\t")
	if err != nil {
		utils.Log.Error("Tag -> json Failed. Error: ", err)
		return err
	}
	err = LevelDbDagInfo.Put([]byte(hash), jsonTag, nil)
	if err != nil {
		utils.Log.Error("Dag Tag -> LevelDB Failed. Error: ", err)
		return err
	}

	if tag.TagType == dag.REQUESTTAG {
		utils.Log.Debug("新的 REQ：", hash)
	}
	if tag.TagType == dag.RESPONSETAG {
		utils.Log.Debug("新的 RESP：", hash)
	}

	// 存储 Tag Table
	err = LevelDbTagTableSaveTag(tag)
	if err != nil {
		utils.Log.Error("Dag Tag -> LevelDB Failed. Error: ", err)
		return err
	}

	//// 如果是新的 Request TAG，判定是否有连接未被连接的TAG
	//if tag.TagType == dag.REQUESTTAG {
	//	UNCONNECTEDLOCK.Lock()
	//	// unConnectedTagMap := GetUnconnectedTagMap()
	//	// utils.Log.Debug("未被连接的TAG有：", unConnectedTagMap)
	//	_, okL := UnConnectedHashMap[tag.PrevHash[0]]
	//	_, okR := UnConnectedHashMap[tag.PrevHash[1]]
	//	if okL == true {
	//		delete(UnConnectedHashMap, tag.PrevHash[0])
	//	}
	//	if okR == true {
	//		delete(UnConnectedHashMap, tag.PrevHash[1])
	//	}
	//	UNCONNECTEDLOCK.Unlock()
	//}

	// 判定是否是可连接的 Tag，如果是可连接的存储到可连接数据库
	if tag.TagType == dag.TOPTAG || tag.TagType == dag.ACKTAG {
		LevelDBSaveAvailableTag(tag)
		if err != nil {
			utils.Log.Error("Save Available Tag Failed. Error: ", err)
		}
	}


	return err
}

func GetTag(hash string) *dag.DagTag {
	jsonTag, err := LevelDbDagInfo.Get([]byte(hash), nil)
	if err != nil {
		utils.Log.Error("Get Tag Failed. Error: ", err)
		return nil
	}
	var tag dag.DagTag
	err = json.Unmarshal(jsonTag, &tag)
	if err != nil {
		utils.Log.Error("Get Tag Failed. Error: ", err)
		return nil
	}
	return &tag
}

// 返回 json
func GetAllDag() []dag.DagTag {
	iter := LevelDbDagInfo.NewIterator(nil, nil)
	var bDagList [][]byte
	var dagList []dag.DagTag
	for iter.Next() {
		// 如果
		value := iter.Value()
		if value == nil {
			break
		}

		var tag dag.DagTag
		bDagList = append(bDagList, value)
		err := json.Unmarshal(value, &tag)
		if err != nil {
			utils.Log.Error(err)
			break
		}
		dagList = append(dagList, tag)
	}
	return dagList
}
