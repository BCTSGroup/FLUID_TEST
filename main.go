package main

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/block"
	"DAG-Exp/src/cfg"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/db_api"
	"DAG-Exp/src/front"
	"DAG-Exp/src/network"
	"DAG-Exp/src/sharecc"

	"DAG-Exp/src/utils"
	"context"
	"encoding/gob"
	"flag"
	"github.com/astaxie/beego/toolbox"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var HttpListenPort string

// gob=go binary  序列化与反序列化，用于从连接中读取信息
func gobInterfaceRegister() {

	gob.Register(block.HandMessage{})
	gob.Register(block.NodeIpAndAccountMessage{})
	gob.Register(block.ConnectMessage{})

	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{}) //this for Transaction.Transaction body

	gob.Register(sharecc.SyncTest{})
	gob.Register(dag.DagTag{})
	gob.Register(dag.TopTagBody{})
	gob.Register(dag.RequestTagBody{})
	gob.Register(dag.ResponseTagBody{})
	gob.Register(dag.AckTagBody{})

	gob.Register(dag.Vote{})
	gob.Register(dag.EpochTag{})
}

func initialize() {
	// 初始化系统参数
	gobInterfaceRegister()
	initMap()
	// 获取启动信息
	getBootInfo()
	// 初始化数据库服务
	initializeDatabase()
	// 初始化节点信息
	initializeAccount()

}

func main() {
	utils.Log.Infof("Code version: 1")
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
	// 启动 http 服务
	go func() {
		log.Fatal(front.HttpApiRun(HttpListenPort))
	}()
	// 启动 p2p 服务 ( 连接网络，同步信息 )
	go p2pServerRun(&Listener, ctx, cfg.GConfig.P2pPort)
	// 启动区块链系统服务 ( 共识服务 )
	go blockchainRun(cfg.GConfig.Bootnodes, cfg.GConfig.P2pPort, &sysDone)

	<-sysDone
	// 关闭数据库服务
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.DbBlock.Close())
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.LevelDbDagInfo.Close())
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.LevelDbTagTable.Close())
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.LevelDBAvailableTag.Close())
	utils.Log.Errorf("Closing DbBlock, Error: %v", db_api.DB_Checksum.Close())
	toolbox.StopTask()
	utils.Log.Infof("All code exit !")
	cancel()

}

func initMap() {
	Network.GConnsIpServerPort = make(map[string]string)
}

func getBootInfo() {
	Network.GLocalIP = cfg.GConfig.Ip
	Network.GPort = cfg.GConfig.P2pPort
	//get blockNode network cfg
	HttpListenPort = cfg.GConfig.HttpPort
}

func initializeDatabase() {
	var err error

	db_api.DbBlock, err = leveldb.OpenFile("./db", nil)
	if err != nil {
		utils.Log.Errorf("db_block error cause: %s", err.Error())
	}
	db_api.LevelDbDagInfo, err = leveldb.OpenFile("./DagInfo", nil)
	if err != nil {
		utils.Log.Errorf("Dag Info DB error cause: %s", err.Error())
	}
	db_api.LevelDBAvailableTag, err = leveldb.OpenFile("./AvailableTag", nil)
	if err != nil {
		utils.Log.Error("Available Tag DB Open Failed. Error: ", err.Error())
	}
	db_api.LevelDbTagTable, err = leveldb.OpenFile("./TagTable", nil)
	if err != nil {
		utils.Log.Errorf("Tag Table error cause: %s", err.Error())
	}
	db_api.DB_Checksum, err = leveldb.OpenFile("./checksum", nil)
	if err != nil {
		utils.Log.Errorf("db_block error cause: %s", err.Error())
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


func p2pServerRun(Listenser *net.Listener, ctx context.Context, listenPort string) {
	*Listenser = Network.NewP2pServer(ctx, listenPort)
}

func blockchainRun(p2pPeerAddress []string, p2pListenPort string, sysDone *chan struct{}) {
	// 作为对等节点去连接种子节点
	utils.Log.Debugf("My Account:%v", account.GAccount)
	isConnectToSeedNodes := Network.ConnectSeedNodes(p2pPeerAddress,
		p2pListenPort,
		*account.GAccount,
		string(account.GAccount.Address))

	if isConnectToSeedNodes == false {
		utils.Log.Errorf("Can't connect to  %s to get chain info", p2pPeerAddress)
		*sysDone <- struct{}{}
		return
	}
	utils.Log.Infof("FINISH CONNECT TO SEED")


}