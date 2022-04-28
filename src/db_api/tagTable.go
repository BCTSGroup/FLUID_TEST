package db_api

import (
	"DAG-Exp/src/dag"
	"DAG-Exp/src/utils"
	"encoding/json"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

var LevelDbTagTable *leveldb.DB
const NOTFOUND = -1

/* 存每个 Tag 对应什么 Type */
// Hash, type
func LevelDbTagTableSaveTag(tag dag.DagTag) error {
	//hash := tag.Hash
	//tagType := strconv.Itoa(tag.TagType)
	//err := LevelDbTagTable.Put([]byte(hash), []byte(tagType), nil)
	//if err != nil {
	//	utils.Log.Error("Dag Tag -> LevelDB Tag Table Failed. Error: ", err)
	//}
	var err error
	if tag.TagType == dag.ACKTAG {
		err = levelDbTagTableSaveTypeList(tag.TagType, tag.Hash)
		if err != nil {
			utils.Log.Error("Dag Tag -> LevelDB Type List Failed. Error: ", err)
		}
	}

	return err
}
func LevelDbTagTableGetTagType(hash string) int {
	bType, err := LevelDbTagTable.Get([]byte(hash), nil)
	if err != nil {
		utils.Log.Error("Get Tag Type Failed. Error: ", err)
		return NOTFOUND
	}
	tagType,err := strconv.Atoi(string(bType))
	if err != nil {
		utils.Log.Error("Get Tag Type Failed. Error: ", err)
		return NOTFOUND
	}
	return tagType
}

/* 存每个 Type 对应了那些 Tag */
// type, []hashlist
func  levelDbTagTableSaveTypeList(tagType int, hash string) error {
	typeHashList, err := LevelDbTagTableGetTypeList(tagType)
	// 当前数据库中存在这个类型的 Tag
	if err == nil {
		// 加入到 List 中，转换成 []byte
		typeHashList = append(typeHashList, hash)
		bHashList, err := json.Marshal(typeHashList)
		if err != nil {
			utils.Log.Error("Save Type List Failed. Error: ", err)
			return err
		}
		// 添加到数据库中
		t := strconv.Itoa(tagType)
		err = LevelDbTagTable.Put([]byte(t), bHashList, nil)
		if err != nil {
			utils.Log.Error("Save Type List Failed. Error: ", err)
			return err
		}
		return nil
	} else {
		// 当前数据库中不存在这个类型的 Tag
		utils.Log.Error("err: 不存在交易类型：", hash, tagType)
		// 创建一个 List，把当前的 hash 添加到 List
		var hashList []string
		hashList = append(hashList, hash)
		bHashList, err := json.Marshal(hashList)
		if err != nil {
			utils.Log.Error("Save Type List Failed. Error: ", err)
			return err
		}
		// 把 List 添加到数据库中
		t := strconv.Itoa(tagType)
		err = LevelDbTagTable.Put([]byte(t), bHashList, nil)
		if err != nil {
			utils.Log.Error("Save Type List Failed. Error: ", err)
			return err
		}
		return nil
	}
}
func LevelDbTagTableGetTypeList(tagType int) ([]string, error) {
	t := strconv.Itoa(tagType)
	bHashList, err := LevelDbTagTable.Get([]byte(t), nil)
	if err != nil {
		utils.Log.Error("Get Tag Type List Failed. Error: ", err)
		return nil, err
	}
	var hashList []string
	err = json.Unmarshal(bHashList, &hashList)
	if err != nil {
		utils.Log.Error("Get Tag Type List Failed, Unmarshal Failed. Error: ", err)
		return nil, err
	}
	return hashList, err
}