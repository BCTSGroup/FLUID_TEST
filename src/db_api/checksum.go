package db_api

import (
	"DAG-Exp/src/utils"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
)

var DB_Checksum *leveldb.DB

func SaveChecksum(h string, prevH []string) {
	var checksum []string
	var cL []string
	var cR []string
	var checksumL string
	var checksumR string
	// checksum[0]自己的校验和 checksum[1]左校验和 checksum[2]右校验和
	for i, v := range prevH {
		if i == 0 && len(v) != 0 {
			// 找前面的校验和
			cL = GetChecksum(v)
		}
		if i == 1 && len(v) != 0 {
			cR = GetChecksum(v)
		}
	}
	if cL != nil {
		checksumL = cL[0]
	}
	if cR != nil {
		checksumR = cR[0]
	}
	s := checksumL + checksumR + h
	ownChecksum, _ := utils.GetHashFromBytes([]byte(s))
	checksum = append(checksum, string(ownChecksum))
	checksum = append(checksum, checksumL)
	checksum = append(checksum, checksumR)
	bChecksum, err := json.Marshal(checksum)

	err = DB_Checksum.Put([]byte(h), bChecksum, nil)

	if err != nil {
		utils.Log.Error("err: 存储校验和失败", err.Error())
	}
}

func GetChecksum(hash string) []string {
	bChecksum, err := DB_Checksum.Get([]byte(hash), nil)
	if err != nil {
		utils.Log.Error("Get Tag Failed. Error: ", err)
		return nil
	}
	var checksum []string
	err = json.Unmarshal(bChecksum, &checksum)
	if err != nil {
		utils.Log.Error("Get Tag Failed. Error: ", err)
		return nil
	}
	return checksum
}
