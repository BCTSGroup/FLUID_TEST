package db_api

import (
	"bfc/src/sharecc"
	"bfc/src/utils"
	"encoding/json"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
)
type systemConfigManager struct {
	dbPtr *leveldb.DB
}

var SysCfgManager = systemConfigManager{
	dbPtr: nil,
}
var ErrSYSCFGDBNotOpen = errors.New("System config pointer is nil.")

func (manager *systemConfigManager) Open(path string) error {
	var err error
	manager.dbPtr, err = leveldb.OpenFile(path, nil)

	if err != nil {
		utils.Log.Errorf("Can't open System config DB in  %s, error： %s", path, err.Error())
		return err
	}

	return nil
}

func (manager *systemConfigManager) Close() error {
	if manager.dbPtr == nil {
		utils.Log.Error("CloseDB, System config haven't been opened.")
		return ErrSYSCFGDBNotOpen
	}
	err := manager.dbPtr.Close()
	return err
}

// 根据 Tx Payload 设置原始数据链的数据共享规则
func (manager *systemConfigManager) SetShareRule(sourceChain string, rule sharecc.JoinPayload) error {
	type DATATYPE    = string
	type SHARE_LEVEL = string
	// (k, v) = (sourceChainID, map[type(string)]level(string))
	value := make(map[DATATYPE]SHARE_LEVEL)
	for _, dataType := range rule.ClassY00{
		value[dataType] = "Y00"
	}
	for _, dataType := range rule.ClassY01{
		value[dataType] = "Y01"
	}
	for _, dataType := range rule.ClassY10{
		value[dataType] = "Y10"
	}
	for _, dataType := range rule.ClassY11{
		value[dataType] = "Y11"
	}

	bKey := []byte(sourceChain)
	bValue, _ := json.Marshal(value)
	err := manager.dbPtr.Put(bKey, bValue, nil)
	if err != nil {
		utils.Log.Errorf("Set failed : %s", err)
	} else {
		utils.Log.Debug(sourceChain + string(bValue))
	}
	return nil
}

// 获取原始数据链的全部共享规则
func (manager *systemConfigManager) GetRule(sourceChain string) map[string]string {
	bRule, err := manager.dbPtr.Get([]byte(sourceChain), nil)
	if err != nil {
		utils.Log.Error(err)
		return nil
	}
	var rule map[string]string
	json.Unmarshal(bRule, &rule)
	return rule
}

// 获取某原始数据链的某类型的共享等级
func (manager *systemConfigManager) GetDataLevel(sourceChain string, dataType string) (string, error) {
	// 从数据库获取配置信息
	bLevelMap, err := manager.dbPtr.Get([]byte(sourceChain), nil)
	if err != nil {
		utils.Log.Error(err)
		return "", err
	}
	// 从配置信息 MAP 中获取定义的共享等级
	type DATATYPE    = string
	type SHARE_LEVEL = string
	var levelMap map[DATATYPE]SHARE_LEVEL
	err = json.Unmarshal(bLevelMap, &levelMap)

	// utils.Log.Debug(string(bLevelMap))

	if err != nil {
		return "", err
	}
	level := levelMap[dataType]

	// utils.Log.Debug(level)

	// 如果不存在，返回错误信息
	if level == "" {
		return "",err
	}
	// 如果存在，返回定义的数据等级
	return level, nil
}