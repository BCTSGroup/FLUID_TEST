package db_api

import (
	"bfc/src/utils"
	"encoding/json"
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

var DbBlock *leveldb.DB





const (
	Existed = iota
	NotFound
	Error
)


func DB_GetBlockByHash(hash string) []byte {
	// 通过 hash 获得对应高度
	height, err := DbBlock.Get([]byte(hash), nil)
	if err != nil {
		utils.Log.Error("没有找到对应的区块高度信息", err)
		return nil
	}
	block, err := DbBlock.Get(height, nil)
	if err != nil {
		utils.Log.Error("没有找到高度对应的区块", err)
		return nil
	}
	return block
}

func Db_insertChainMem(jsonGChainMem []byte) {
	//序列化,把结构体变为json
	err := DbBlock.Put([]byte(utils.BLOCK_MEM_STRING), jsonGChainMem, nil)
	if err != nil {
		utils.Log.Errorf("json_gChainMem db store failed: %s", err.Error())
	}
}

func Db_GetChainMem() ([]byte, error) {
	ChainMem_json, err := DbBlock.Get([]byte(utils.BLOCK_MEM_STRING), nil)
	return ChainMem_json, err
}

func Db_insertBlockJson(Height int64, Hash string, blockJson []byte) {

	DbBlock.Put([]byte(Hash), []byte(utils.BLOCK_STRING+string(Height)), nil)

	err := DbBlock.Put([]byte(utils.BLOCK_STRING+string(Height)), blockJson, nil)
	if err != nil {
		utils.Log.Errorf("BlockJson db store failed: %s", err.Error())
	}

}

func Db_getBlock(Height int64) ([]byte, error) {
	blockdata, err := DbBlock.Get([]byte(utils.BLOCK_STRING+string(Height)), nil)
	if err == nil {
		return blockdata, nil
	} else {
		utils.Log.Errorf("get block from chain failed, Height:%d", Height)
		return nil, errors.New("get block from chain failed")
	}
}

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

/*
	[function]: Get full nodes' address
	[output]: Json of address slicec
*/
func Db_getAllAddress() ([]byte, error) {
	jsonAddresses, err := DbBlock.Get([]byte(utils.ALL_ADDRESS_STRING), nil)
	if err != nil {
		utils.Log.Errorf("get all address failed: %s", err.Error())
		return nil, errors.New("get all address from db failed")
	} else {
		return jsonAddresses, nil
	}
}

/*
	[function]: Get full nodes' address
	[output]: address slice
*/
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

/*
	[function]: Store all full nodes' address
	[input]: Addresses needed to Insert
*/
func Db_updateSuperAddress(Address string) error {
	AddressSlice, _ := Db_getSuperAddressAsSlice()
	AddressSlice = append(AddressSlice, Address)
	jsonAddressSlice, err := json.Marshal(AddressSlice)
	if err != nil {
		utils.Log.Errorf("SuperAddressSummary serilized failed: %s", err.Error())
		return err
	}
	err = DbBlock.Put([]byte(utils.SUPER_ADDRESS_STRING), jsonAddressSlice, nil)
	if err != nil {
		utils.Log.Errorf("SuperAddressSummary db update failed: %s", err.Error())
		return err
	}
	return nil
}

/*
	[function]: Get full nodes' address
	[output]: Json of address slicec
*/
func Db_getSuperAddress() ([]byte, error) {
	jsonAddresses, err := DbBlock.Get([]byte(utils.SUPER_ADDRESS_STRING), nil)
	if err != nil {
		utils.Log.Errorf("get super address failed: %s", err.Error())
		return nil, errors.New("get super address from db failed")
	} else {
		return jsonAddresses, nil
	}
}

/*
	[function]: Get full nodes' address
	[output]: address slice
*/
func Db_getSuperAddressAsSlice() ([]string, error) {
	jsonAddresses, err := Db_getSuperAddress()
	if err != nil {
		utils.Log.Errorf("get super address json in getting slice of address failed: %s", err.Error())
		return nil, err
	}
	var addressSlice []string
	err = json.Unmarshal(jsonAddresses, &addressSlice)
	if err != nil {
		utils.Log.Errorf("Super ddress unmarshal failed: %s", err.Error())
		return nil, err
	}
	return addressSlice, nil
}

/*
	[function]: Get blockchain information in db
	[output]:	Json of Blockchain info
*/
func Db_getBlockChainSummary() ([]byte, error) {
	jsonSummary, err := DbBlock.Get([]byte(utils.BLOCK_MEM_SUM_STRING), nil)
	if err != nil {
		utils.Log.Errorf("get BlockChain summary failed: %s", err.Error())
		return nil, errors.New("get BlockChainSummary from db failed")
	} else {
		return jsonSummary, nil
	}
}
