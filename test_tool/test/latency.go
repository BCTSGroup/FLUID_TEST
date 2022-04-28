package test

import (
	"DAG-Exp/src/utils"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

type DagTag struct {
	TimeStamp   int64           `json:"timeStamp"`
	Depth       int				`json:"depth"`
	Hash 		string			`json:"hash"`
	PrevHash 	[]string			`json:"prevHash"`
	Path        string			`json:"path"`
	TagType 	int				`json:"tagType"`
	Body 		interface{} 	`json:"body"`
	Miner 		string			`json:"miner"`
	Checkpoint 	string			`json:"signature"`
}

var db *leveldb.DB
var err error
var arr []string
var check []string
func init()  {
	db, err = leveldb.OpenFile("./db", nil)
	if err != nil {
		utils.Log.Errorf("db_block error cause: %s", err.Error())
	}
	l := make([]DagTag, 1000)
	for i, _ := range l {
		l[i] = DagTag{
			TimeStamp: time.Now().UnixNano(),
			Depth:     i,
			Hash:      "",
			PrevHash:  nil,
			Path:      "",
			TagType:   3,
			Body:      nil,
			Miner:     "ABXSDIDWAAGFSUYFDS",
			Checkpoint: "IUISDANFKSNFEURUYD",
		}
		l[i].Hash = calculateTagHash(l[i])
		bTag, _ := json.MarshalIndent(l[i],"","\t")
		_ = db.Put([]byte(l[i].Hash), bTag, nil)
		arr = append(arr, l[i].Hash)
		check = append(check, l[i].Checkpoint)
	}
}

func CalculateVRFLatency() {
	startTime := time.Now()
	for i, v := range arr {
		bTag, _ := db.Get([]byte(v), nil)
		var tag DagTag
		json.Unmarshal(bTag, &tag)

		if check[i] == tag.Checkpoint {
			fmt.Println("Pass.")
			continue
		} else {
			fmt.Println("Error vrf.")
		}
	}

	elapsedTime := time.Since(startTime)  // duration in ms
	fmt.Println("Segment finished in %d ms", elapsedTime)
	defer db.Close()
}

func calculateTagHash(tag DagTag) string{
	bTag, _ := json.Marshal(tag)
	bHash, _ := utils.GetHashFromBytes(bTag)
	return string(utils.Base58Encode(bHash))
}