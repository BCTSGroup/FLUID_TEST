package main

import (
	"bfc/src/cfg"
	"bfc/src/front"
	"bfc/src/sharecc"
	"bfc/src/test"
	"bfc/src/utils/http_utils"
	"bfc/src/utils/pbft_utils"
	"context"
	"encoding/gob"
	"encoding/json"
	"flag"
	"github.com/astaxie/beego/toolbox"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bfc/src/account"
	"bfc/src/block"
	"bfc/src/db_api"
	"bfc/src/utils"
)

var HttpListenPort string

// gob=go binary  序列化与反序列化，用于从连接中读取信息
func gobInterfaceRegister() {
	gob.Register(block.Block{})
	gob.Register(block.Transaction{})
	gob.Register(block.StageMessage{})
	gob.Register(block.HandMessage{})
	gob.Register(block.NodeIpAndAccountMessage{})
	gob.Register(block.ConnectMessage{})
	gob.Register(block.VoteRequest{})
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{}) //this for Transaction.Transaction body
	gob.Register(http_utils.Contract{})
	gob.Register(http_utils.TransferContract{})
	gob.Register(http_utils.TransferContractPayload{})
	gob.Register(sharecc.JoinRequest{})
	gob.Register(sharecc.JoinPayload{})
	gob.Register(sharecc.SubmitRequest{})
	gob.Register(sharecc.SubmitPayload{})
	gob.Register(sharecc.ValidData{})
	gob.Register(sharecc.TxSharingData{})
	gob.Register(sharecc.GetSharedDataRequest{})
	gob.Register(sharecc.GetResponsePayload{})
	gob.Register(sharecc.GetResponse{})
	gob.Register(sharecc.GetResponsePayload{})
	gob.Register(sharecc.DataSharingRecord{})
	gob.Register(sharecc.SharedData{})
	gob.Register(sharecc.MerklePath{})
	gob.Register(test.TestTx{})
	gob.Register(test.TXRequestData{})
	gob.Register(test.TXResponseData{})

}

func initialize() {
	block.CONSENSUS_MAP = make(map[string]bool)
	// 初始化系统参数
	gobInterfaceRegister()
	initMap()
	// 获取启动信息
	getBootInfo()
	// 初始化数据库服务
	initializeDatabase()
	// 初始化节点信息
	initializeAccount()
	// 检查是否发放补助
	checkSubsidized()
	// 获取区块链信息
	getChainInfo()
}

func main() {
	utils.Log.Infof("Code version:3")
	ctx, cancel := context.WithCancel(context.Background())
	sysDone := make(chan struct{}, 1)
	sigCh := make(chan os.Signal, 1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		utils.Log.Info("Receive signal SIGINT/SIGTERM. Terminating...")
		sysDone <- struct{}{}
	}()
	flag.Parse()
	// 系统初始化
	initialize()
	var Listener net.Listener
	// 启动 http 合约服务 ( 同链码 )
	go func() {
		log.Fatal(front.HttpApiRun(HttpListenPort))
	}()
	// 启动 p2p 服务 ( 连接网络，同步信息 )
	go P2pServerRun(&Listener, ctx, cfg.GConfig.P2pPort)
	// 启动区块链系统服务 ( 共识服务 )
	go BlockchainRun(cfg.GConfig.Bootnodes, block.GChainMem.Height, cfg.GConfig.P2pPort, &sysDone)

	<-sysDone
	// 关闭数据库服务
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.DbBlock.Close())
	utils.Log.Errorf("Closing DB SHARE, Error: %v", db_api.ShareManager.Close())
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.SysCfgManager.Close())
	toolbox.StopTask()
	utils.Log.Infof("All code exit !")
	cancel()

}

func initMap() {
	block.GConnsIpServerPort = make(map[string]string)
	//Network.GAddressConns = make(map[string]Network.Conns)
	pbft_utils.PBFTVoteAuthPrepareMap = make(map[string]int, 0)
	pbft_utils.PBFTVoteAuthCommitMap = make(map[string]int, 0)
	pbft_utils.PBFTAuthPrePrepareMap = make(map[string][]byte, 0)
	pbft_utils.PBFTAuthPrepareMap = make(map[string][]byte, 0)
	pbft_utils.PBFTAuthCommitMap = make(map[string][]byte, 0)
	//交易池初始化
	block.GTransactionPool.TransactionMap = make(map[string]block.Transaction)
}

func getBootInfo() {
	block.GLocalIP = cfg.GConfig.Ip
	block.GPort = cfg.GConfig.P2pPort
	//get blockNode network cfg
	HttpListenPort = cfg.GConfig.HttpPort
}

func initializeDatabase() {
	var err error

	db_api.DbBlock, err = leveldb.OpenFile("./db", nil)
	if err != nil {
		utils.Log.Errorf("db_block error cause: %s", err.Error())
	}
	err = db_api.SysCfgManager.Open("./shareRule")
	if err != nil {
		utils.Log.Errorf("System config database open failed.")
	}
	err = db_api.ShareManager.Open("./shareTable")
	if err != nil {
		utils.Log.Errorf("Sharing data database open failed.")
	}

}

func initializeAccount() {
	if account.CheckPrivatePemExistance() && account.CheckPublicPemExistance() {
		// if the account folder has publicPem which mean it is the first block
		// then
		account.GAccount, _ = account.InitAccountFromLocalFile()
		block.GMyAccountAddress = string(account.GAccount.Address)
	} else {
		account.GAccount, _ = account.NewAccount()
		block.GMyAccountAddress = string(account.GAccount.Address)
	}
}

func checkSubsidized() {
	for _, producer := range block.GChainMem.ProduceAccount {
		if producer == string(account.GAccount.Address) && account.GAccount.IsSubsidized == false {
			//fixme To change!
			account.GAccount.IsSubsidized = true
			account.GAccount.Balance = 1e20
			break
		}
	}
	_, err := account.SaveAccountInfoToDb(*account.GAccount)
	if err != nil {
		utils.Log.Errorf("Save Init account failed! Error: %s", err.Error())
	}
}

func getChainInfo() {
	ChainMemJson, err := db_api.Db_GetChainMem()
	if err == nil {
		//if the block_mem_string get successfully
		//which mean this node has restart
		//反序列化，把json解析到结构体里面
		_ = json.Unmarshal(ChainMemJson, &block.GChainMem)
		utils.Log.Infof("block: %v", block.GChainMem)
	} else {
		//the first time node start
		block.GChainMem.Height = 0
		block.GChainMem.Hash = "0000"
		block.GChainMem.Timestamp = time.Now().Unix()
		block.GChainMem.PrevBlockHash = "0000"
		// TODO : Set the block producer, it must be changed to be related to the vote results

		block.GChainMem.ProduceAccount = cfg.GConfig.ProducerAddress

		if len(cfg.GConfig.Ip) == 0 {
			block.GChainMem.NodeType = block.BLOCK_GENESIS
		} else {
			block.GChainMem.NodeType = block.BLOCK_LIGHT
		}

		jsonGChainMem, err := json.Marshal(block.GChainMem)
		if err != nil {
			utils.Log.Errorf("gChainMem serilized failed: %s", err.Error())
		}
		db_api.Db_insertChainMem(jsonGChainMem)
	}
}

func P2pServerRun(Listenser *net.Listener, ctx context.Context, listenPort string) {
	*Listenser = block.NewP2pServer(ctx, listenPort)
}

func BlockchainRun(p2pPeerAddress []string, Height int64, p2pListenPort string, sysDone *chan struct{}) {
	//if len(p2p_peer_address)!= 0{
	//	for false == Network.ConnectSeedNodes(p2p_peer_address, Height, p2p_listen_port){}
	//}
	//var ProducerAccountRWlock sync.RWMutex

	utils.Log.Debugf("My Account:%v", account.GAccount)
	isConnectToSeedNodes := block.ConnectSeedNodes(p2pPeerAddress,
		Height,
		p2pListenPort,
		*account.GAccount,
		string(account.GAccount.Address))

	if isConnectToSeedNodes == false {
		utils.Log.Errorf("Can't connect to  %s to get chain info", p2pPeerAddress)
		*sysDone <- struct{}{}
		return
	}

	utils.Log.Infof("FINISH CONNECT TO SEED")

	//现在全部节点都是超级节点
	//创世节点直接出块，如果不是创世节点还要判断是否有出块权限，是否是出块时机
	//在出块的超级节点中，并且本轮次出块才可以出块
	block.GChainMemRWlock.RLock()
	NodeType := block.GChainMem.NodeType
	ProduceAccount := block.GChainMem.ProduceAccount
	lastBlockTime := block.GChainMem.Timestamp
	utils.Log.Debugf("Produce accounts(GChainMem) : %s" , block.GChainMem.ProduceAccount)
	utils.Log.Debugf("Node type： %d, Produce Account :  %s", NodeType, ProduceAccount)
	block.GChainMemRWlock.RUnlock()

	//该定时器用来统计投票结果，决定出块节点
	go block.DPOSTimer()

	var ticker = time.NewTicker(time.Millisecond * 250)
	block.P2PBroadcastToSuperCallback = block.P2pBroadcastTransaction

	for {
		select {
		case <-ticker.C:
			currentSlot := block.GetSlotNumber(0)
			lastSlot := block.GetSlotNumber(utils.GetEpochTime(lastBlockTime))

			//相同的秒数说明刚刚出的块距离现在还没过去一秒
			//如果两个slot相同就等待100毫秒,下面的代码用于在一轮出块中轮回
			if currentSlot == lastSlot {
				//time.Sleep(time.Millisecond * 100)
				continue
			}

			delegateID := currentSlot % cfg.NProducers

			var CanProduce = false
			var MyProduceIndex = 0

			for index, str := range ProduceAccount {
				if str == block.GMyAccountAddress {
					CanProduce = true
					MyProduceIndex = index
				}
			}


			//timer interrupt
			//TO detect the vote result
			//if block.GDPOSFlag == true {
			//	block.ResetVoteMessage()
			//	block.GDPOSFlag = false
			//
			//	// 5 producers
			//	ProduceAccount := make([]string, len(cfg.GConfig.ProducerAddress))
			//	copy(ProduceAccount, cfg.GConfig.ProducerAddress)
			//
			//	if len(block.GVotedAddressesInPair) > 0 {
			//		if len(block.GVotedAddressesInPair) >= len(ProduceAccount) {
			//			for k, v := range block.GVotedAddressesInPair {
			//				if k >= len(ProduceAccount)-1 {
			//					break
			//				}
			//				ProduceAccount[k] = v.Key
			//
			//			}
			//		} else {
			//			for k, v := range block.GVotedAddressesInPair {
			//				ProduceAccount[k] = v.Key
			//			}
			//		}
			//	}
			//	block.GChainMem.ProduceAccount = ProduceAccount
			//	jsonGChainMem, err := json.Marshal(block.GChainMem)
			//	if err != nil {
			//		utils.Log.Errorf("gChainMem serilized failed: %s", err.Error())
			//	}
			//	block.GChainMemRWlock.Lock()
			//	db_api.Db_insertChainMem(jsonGChainMem)
			//	block.GChainMemRWlock.Unlock()
			//
			//}
			block.CONSENSUS_MAP_LOCK.RLock()
			isAllPass := true
			for _,v := range block.CONSENSUS_MAP {
				if v == false {
					isAllPass = false
					break
				}
			}
			block.CONSENSUS_MAP_LOCK.RUnlock()

			if CanProduce && isAllPass {
				//查询自己的出块顺序是不是当前出块节点
				if int64(MyProduceIndex) == delegateID {
					// 查看交易池中是否有足够交易可以用来打包
					block.GTransactionPoolRWlock.RLock()
					txNumInTxPool := len(block.GTransactionPool.TransactionMap)
					block.GTransactionPoolRWlock.RUnlock()
					utils.Log.Debug("待出块，检查交易池数据...")
					if txNumInTxPool == 0 {
						utils.Log.Infof("交易池为空，本轮轮空")
						time.Sleep(time.Second * time.Duration(block.BLOCK_PRODUCE_INTERVAL))
						continue
					}

					// 提议新区块
					utils.Log.Debug("开始提议新区块...")
					newBlock := block.NewBlock()
					utils.Log.Debug("新区块提议完成，内含交易：",newBlock.Head.TxCount)
					// utils.Log.Debug(newBlock)
					//BlockJson, err := json.Marshal(newBlock)
					//if err != nil {
					//	utils.Log.Errorf("Block serilized failed: %s", err.Error())
					//}
					//blockReceive := block.Block{}
					//_ = json.Unmarshal(BlockJson, &blockReceive)
					_ = block.PreExecuteBlock(newBlock, true)

					go block.P2pBroadcastPrePrepare(newBlock)
					//go Network.P2pBroadcastToAllNodesBlock(BlockJson)
					time.Sleep(time.Second * time.Duration(block.BLOCK_PRODUCE_INTERVAL))
				} else {
					//不能出块则睡眠500毫秒
					//time.Sleep(time.Second * 500)
					continue
				}
			}
		}
	}

}