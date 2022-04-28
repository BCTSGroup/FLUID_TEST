package cfg

import (
	"bfc/src/utils"
	"os"
)

var GConfig Config
var NProducers int64 // number of producers / super nodes.

type RnConfig struct {
	Rnid    string `json:"rnid"`
	Address string `json:"address"`
}

type Config struct {
	Ip                   string     `json:"ip"`
	P2pPort              string     `json:"p2pPort"`
	HttpPort             string     `json:"httpPort"`
	Bootnodes            []string   `json:"bootNodes"`
	ProducerAddress      []string   `json:"ProducerAddress"`
}

/*
 * init global variable GConfig from local file "boot.json"
 * this func is called only once no matter how many times this package is imported
 */
func init() {
	// Unmarshal local json file
	JsonParse := utils.NewJsonStruct()
	err := JsonParse.Load("boot.json", &GConfig)

	// load file error, exit
	if err != nil {
		utils.Log.Fatalf("Load boot.json error:  %s", err.Error())
		os.Exit(1)
	}

	//if GConfig.ProfitAllocationTime == "" {
	//	GConfig.ProfitAllocationTime = "0 0 2 * * *"
	//}
	utils.Log.Debugf("CFG produce account : %s",GConfig.ProducerAddress)

	NProducers = int64(len(GConfig.ProducerAddress))
}
