package db_api

import (
	"bfc/src/sharecc"
	"bfc/src/utils"
	"encoding/json"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
)

type shareDataManager struct {
	dbPtr *leveldb.DB
}

var ShareManager = shareDataManager{
	dbPtr: nil,
}

var ErrSHAREDBNotOpen = errors.New("SHARE database pointer is nil")



func (manager *shareDataManager) Open(path string) error {
	var err error
	manager.dbPtr, err = leveldb.OpenFile(path, nil)

	if err != nil {
		utils.Log.Errorf("Can't open SHARE db in  %s, error： %s", path, err.Error())
		return err
	}

	return nil
}

func (manager *shareDataManager) Close() error {
	if manager.dbPtr == nil {
		utils.Log.Error("CloseDB, SHARE database haven't been opened.")
		return ErrSHAREDBNotOpen
	}
	err := manager.dbPtr.Close()
	return err
}

func (manager *shareDataManager) TempSaving(from string, data sharecc.TxSharingData) error {
	key := data.Key
	value := sharecc.SharedData{
		DataType:  data.DataType,
		Value:     data.Value,
		ValidTime: data.TimeValid,
	}
	// 查询该类型的存储等级
	level, err := SysCfgManager.GetDataLevel(from, data.DataType)
	if err != nil {
		utils.Log.Error("没有找到数据等级记录",err)
		return err
	}
	value.Level = strings.Replace(level, "Y", "X", -1)

	bValue, _ := json.Marshal(value)

	utils.Log.Debug(key)
	utils.Log.Debug(string(bValue))

	err = manager.dbPtr.Put([]byte(key), bValue, nil)
	if err != nil {
		utils.Log.Error("临时数据存储失败，", err)
		return err
	}
	return nil
}

func (manager *shareDataManager) SaveSharingData(blockHash string,data sharecc.TxSharingData) error {
	key := data.Key

	utils.Log.Debug(key)

	tempValue, err := manager.dbPtr.Get([]byte(key), nil)
	if err != nil {
		utils.Log.Error("读取临时数据失败",err)
		return err
	}

	utils.Log.Debug(string(tempValue))

	var value sharecc.SharedData
	json.Unmarshal(tempValue, &value)
	// 1. 更改存储等级 X -> Y
	value.Level = strings.Replace(value.Level, "X", "Y", -1)

	utils.Log.Debug(value.Level)

	// 2. 添加验证索引(Block Hash)
	value.BlockHash = blockHash

	// 3. 将更改后的共享数据存入数据库
	bValue,_ := json.Marshal(value)
	err = manager.dbPtr.Put([]byte(key), bValue, nil)
	if err != nil {
		utils.Log.Error("共享数据存储失败",err)
		return err
	}

	return nil
}

func (manager *shareDataManager) GetSharedData(key string) *sharecc.SharedData{
	bSharedData, err := manager.dbPtr.Get([]byte(key), nil)
	if err != nil {
		utils.Log.Error("没有找到共享数据",err)
		return nil
	}
	sharedData := &sharecc.SharedData{}
	json.Unmarshal(bSharedData, sharedData)
	return sharedData
}