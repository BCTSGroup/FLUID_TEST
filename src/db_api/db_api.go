package db_api

import (
	"DAG-Exp/src/utils"
	"encoding/json"
	"errors"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

var DbBlock *leveldb.DB





const (
	Existed = iota
	NotFound
	Error
)


// 账户相关的
func Db_insertAccount(jsonAccount []byte, Address string) error {
	err := DbBlock.Put([]byte(utils.ACCOUNT_STRING+string(Address)), jsonAccount, nil)
	if err != nil {
		utils.Log.Errorf("json_account db store failed: %s", err.Error())
	}
	return err
}
func Db_getAccount(Address string) ([]byte, error) {
	jsonAccountinfo, err := DbBlock.Get([]byte(utils.ACCOUNT_STRING+ Address), nil)
	if err == nil {
		return jsonAccountinfo, err
	} else {
		utils.Log.Errorf("get Account Information failed: %s", err.Error())
		return nil, errors.New("get Account from db failed")
	}
}
func Db_updateAllAddress(Address string) error {
	AddressSlice, _ := Db_getAllAddressAsSlice()
	for k, v := range AddressSlice {
		if v == Address {
			AddressSlice = append(AddressSlice[0:k], AddressSlice[k+1:]...)
		}
	}
	AddressSlice = append(AddressSlice, Address)
	jsonAddressSlice, err := json.Marshal(AddressSlice)
	if err != nil {
		utils.Log.Errorf("AllAddressSummary serilized failed: %s", err.Error())
		return err
	} else {
		err = DbBlock.Put([]byte(utils.ALL_ADDRESS_STRING), jsonAddressSlice, nil)
		if err != nil {
			utils.Log.Errorf("AllAddressSummary db update failed: %s", err.Error())
			return err
		}
		return nil
	}
}
func Db_getAllAddress() ([]byte, error) {
	jsonAddresses, err := DbBlock.Get([]byte(utils.ALL_ADDRESS_STRING), nil)
	if err != nil {
		utils.Log.Errorf("get all address failed: %s", err.Error())
		return nil, errors.New("get all address from db failed")
	} else {
		return jsonAddresses, nil
	}
}
func Db_getAllAddressAsSlice() ([]string, error) {
	jsonAddresses, err := Db_getAllAddress()
	if err != nil {
		utils.Log.Errorf("get all address json in getting slice of address failed: %s", err.Error())
		return nil, err
	} else {
		var addressSlice []string
		err = json.Unmarshal(jsonAddresses, &addressSlice)
		if err != nil {
			utils.Log.Errorf("all Address unmarshal failed: %s", err.Error())
			return nil, err
		} else {
			return addressSlice, nil
		}

	}

}


// DAG test

func CalculateLatency(hash string) int64{
	ackTag := GetTag(hash)
	preHash := ackTag.PrevHash
	endT := ackTag.TimeStamp

	respTag := GetTag(preHash[0])
	reqTag  := GetTag(respTag.PrevHash[0])
	reqT := reqTag.TimeStamp

	latency := ( endT - reqT ) / int64(time.Microsecond)
	return latency
}





