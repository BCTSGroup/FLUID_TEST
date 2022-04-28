package block

import (
	"bfc/src/cfg"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"bfc/src/account"
	"bfc/src/db_api"
	"bfc/src/utils"
)

type Pair struct {
	Key   string
	Value uint64
}

var CONSENSUS_MAP map[string]bool
var CONSENSUS_MAP_LOCK sync.RWMutex


type PairList []Pair

var GVotedAddressesInPair PairList
var GDPOSFlag bool

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

func sortMapByValue(m map[string]uint64) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
	}
	sort.Sort(p)
	return p
}

func QueryDPOSVotes() {
	//开始统计投票结果
	VotedAddresses := make(map[string]uint64, 30)
	//获取所有的节点的地址，用于记录他们投票地址
	NodeAddresses, _ := db_api.Db_getAllAddressAsSlice()
	var NodeAccount account.AccountInDb
	sumCount, votingCount := 0, 0
	for _, address := range NodeAddresses {
		NodeAccount, _ = account.GetAccountAsStructureFromDB(address)
		sumCount++
		//util.Log("[", utils.GetUTCTime(), "]  Querying Dpos Votes  :  ", address, " : pledge : ", NodeAccount.Pledge)
		if NodeAccount.Votes > 0 {
			VotedAddresses[address] = NodeAccount.Votes
		}
		if NodeAccount.Pledge > 0 {
			votingCount++
		}
	}
	//util.Log("[", utils.GetUTCTime(), "]  Querying Dpos Votes  :  ", votingCount, "Vote List : ", VotedAddresses)

	//fixme: whether need a requirement of total amount?
	if votingCount >= sumCount*2/3 {
		GVotedAddressesInPair = sortMapByValue(VotedAddresses)
		utils.Log.Infof(" Querying Dpos Votes  :  use Dpos!, %v", GVotedAddressesInPair)

		return
	} else {
		//util.Log("[", utils.GetUTCTime(), "]  Querying Dpos Votes  :  use Fixed!")
		return
	}
}

func DPOSTimer() {
	var INTERVAL = BLOCK_PRODUCE_INTERVAL
	var NodeROUND = cfg.NProducers

	//创建定时器并设置定时时间
	Timer := time.NewTimer(time.Duration(NodeROUND*INTERVAL) * time.Second)

	//循环监听定时器
	for {
		select {
		case <-Timer.C:
			//中断标志位，超时后输出
			QueryDPOSVotes()
			//处理pbft等过程内存中存下来的无用记录
			StorageClearFunc()
			//相当于一个互斥锁
			GDPOSFlag = true
			//超时后重置定时器
			Timer.Reset(time.Duration(NodeROUND*INTERVAL) * time.Second)
		}
		//util.Log("[", utils.GetUTCTime(), "]  BlockRunning  :  timeup! ")

	}
}

func ResetVoteMessage() {
	NodeAddresses, _ := db_api.Db_getAllAddressAsSlice()
	var NodeAccount account.AccountInDb
	for _, address := range NodeAddresses {
		NodeAccount, _ = account.GetAccountAsStructureFromDB(address)
		NodeAccount.Balance += float64(NodeAccount.Pledge)
		NodeAccount.Pledge = 0
		NodeAccount.Votes = 0
		jsonAccount, err := json.Marshal(NodeAccount)
		if err != nil {
			utils.Log.Errorf("Account Info serilized failed: %s", err.Error())
		}
		err = db_api.Db_insertAccount(jsonAccount, address)
		return
	}
}

func StorageClearFunc() {
	//PBFT过程中的无用记录

}
