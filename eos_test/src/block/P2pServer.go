package block

import (
	"bfc/src/sharecc"
	"bfc/src/utils/pbft_utils"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"bfc/src/account"
	"bfc/src/db_api"
	"bfc/src/utils"
)

var GPort string //for single host
var GLocalIP string

/**********************************************************************************************************************
												   新建 P2P 服务
**********************************************************************************************************************/
func NewP2pServer(ctx context.Context, listenPort string) net.Listener {
	//listener, err := net.Listen("tcp", "10.0.2.15:"+strconv.FormatInt(int64(listenPort+node.ID), 10))
	utils.Log.Info("Setting up NewP2pServer  :  p2p listen port:" + GLocalIP + ":" + listenPort)
	listener, err := net.Listen("tcp", GLocalIP+":"+listenPort)

	if err != nil {
		log.Println("Setting up NewP2pServer  :  NewServer Failed")
		return nil
	}

	go func(ctx context.Context, listener net.Listener) {
	EndListener:
		for {

			conn, err := listener.Accept()

			if err != nil {
				log.Println("Running NewP2pServer  :  Accept Failed %s", err.Error())
			}

			newConns := Conns{conn: conn, PtrEncoder: gob.NewEncoder(conn), RemoteAddress: "some thing connect in"}

			GConnsRWlock.Lock()
			GConns = append(GConns, newConns)
			//大概是路由表吧？GConnsIpServerPort
			GConnsIpServerPort[conn.RemoteAddr().String()] = "None"
			GConnsRWlock.Unlock()

			//创建connection的解码器，并传到P2pHandleConnection中进行解码
			dec := gob.NewDecoder(conn)
			utils.Log.Info("Running up NewP2pServer  :  receive a conn:", conn.RemoteAddr())
			go P2pHandleConnection(newConns, dec)

			ifCloseSuccessfully := true
			select {
			case <-ctx.Done():
				for _, conn := range GConns {
					if conn == (Conns{}) {
						continue
					}
					if err := conn.conn.Close(); err != nil {
						ifCloseSuccessfully = false
						utils.Log.Errorf("Running up NewP2pServer  :  close connections error!! %s", err.Error())
					}
				}
				if err := listener.Close(); err != nil {
					ifCloseSuccessfully = false
					utils.Log.Errorf("Running up NewP2pServer  :  close listener error!! %s", err.Error())
				}
				if ifCloseSuccessfully {
					utils.Log.Debug("End all connections and listener")
				}
				// jump out the loop
				break EndListener
			default:
			}

			//time.Sleep(time.Millisecond * 100)
		}
	}(ctx, listener)

	return listener
}

/**********************************************************************************************************************
             								      P2P 消息处理路由
**********************************************************************************************************************/
func ProcessMessage(msg *Message, conn Conns) {
	switch msg.Type {
	case MessageTypeInit:
		handleInitMessage(msg, conn)
	case MessageTypeHandshake:
		handleHandshakeMessage(msg, conn)
	case MessageTypeNodeIp:
		handleNodeIpMessage(msg, conn)
	case MessageTypeTransaction:
		handleTransactionMessage(msg, conn)
	case MessageTypeRenovateBlock:
		handleBlockJsonRenovateMessage(msg, conn)
	case MessageTypeSyncBlockJson:
		handleSyncBlockJsonMessage(msg, conn)
	case MessageTypeBlockJson:
		handleBlockJsonMessage(msg, conn)
	case MessageTypePrePrepare:
		handlePbftPrePrepareMessage(msg, conn)
	case MessageTypePrepare:
		handlePbftPrepareMessage(msg, conn)
	case MessageTypeCommit:
		handlePbftCommitMessage(msg, conn)
	case MessageTypeConnect:
		handleConnectWithAccountMessage(msg, conn)
	case MessageTypeSubmitSharingData:
		handleSubmitSharedData(msg)
	case MessageTypeGetSharedData:
		handleGetSharedData(msg)
	default:
		//Pbft.ProcessStageMessage(msg)
	}
}

func handleSubmitSharedData(msg *Message) {
	bRequest, _ := json.Marshal(msg.Body)
	var submitRequest sharecc.SubmitRequest
	err := json.Unmarshal(bRequest, &submitRequest)
	if err != nil {
		utils.Log.Error("Submit Request 解析失败！")
		return
	}
	isExecd := Contract_ReceiveSubmitData(submitRequest)
	if !isExecd {
		utils.Log.Error("执行提交处理合约失败！")
		return
	}
	return
}

func handleGetSharedData(msg *Message){
	bRequest, _ := json.Marshal(msg.Body)
	var getDataRequest sharecc.GetSharedDataRequest
	err := json.Unmarshal(bRequest, &getDataRequest)
	if err != nil {
		utils.Log.Error("Submit Request 解析失败！")
		return
	}
	Contract_DataSharingRecord(getDataRequest)
}

func handleBlockJsonMessage(msg *Message, conn Conns) {
	//get the message
	messageByte, _ := json.Marshal(msg.Body)
	messageHash, _ := utils.GetHashFromBytes(messageByte)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, messageHash, messageSignature)
	if !isVerify {
		//if verify failed , end the processing
		utils.Log.Errorf("Handling received block json verify failed signature not true, msg: %s", msg)
		return
	}

	// receive a block from blockchain
	utils.Log.Infof("receive blockJsonMessage from : %s", conn.conn.RemoteAddr().String())
	blockJson := msg.Body.([]byte)
	// utils.Log.Infof("blockJson: %s", string(blockJson))

	//json解析到结构体里面
	blockReceive := Block{}
	err := json.Unmarshal(blockJson, &blockReceive)
	if err == nil {
		if ValidateBlock(&blockReceive) {

			//keyVerify := strconv.FormatInt(blockReceive.Height, 16) + "_" + string(pubKeyPeer) + "_" + blockReceive.Hash
			//delete the useless in map
			//utils.Log.Debug("删除 Vote Auth PBFT Prepare Map...")
			//pbft_utils.PBFTVoteAuthPrepareMapLock.Lock()
			//delete(pbft_utils.PBFTVoteAuthPrepareMap, blockReceive.Head.Hash)
			//pbft_utils.PBFTVoteAuthPrepareMapLock.Unlock()
			//
			//utils.Log.Debug("删除 Vote Auth PBFT Commit Map...")
			//pbft_utils.PBFTVoteAuthCommitMapLock.Lock()
			//delete(pbft_utils.PBFTVoteAuthCommitMap, blockReceive.Head.Hash)
			//pbft_utils.PBFTVoteAuthCommitMapLock.Unlock()
			//
			//utils.Log.Debug("删除 PBFT Auth Pre-prepare Map...")
			//pbft_utils.PBFTAuthPrePrepareMapRWLock.Lock()
			//for deleteKey, _ := range pbft_utils.PBFTAuthPrePrepareMap {
			//	splitResult := strings.Split(deleteKey, "_")
			//	if len(splitResult) > 1 {
			//		if splitResult[2] == blockReceive.Head.Hash {
			//			delete(pbft_utils.PBFTAuthPrePrepareMap, deleteKey)
			//			break
			//		}
			//	}
			//
			//}
			//delete(pbft_utils.PBFTAuthPrePrepareMap, blockReceive.Head.Hash)
			//pbft_utils.PBFTAuthPrePrepareMapRWLock.Unlock()
			//
			//utils.Log.Debug("删除 PBFT Auth prepare Map...")
			//pbft_utils.PBFTAuthPrepareMapRWLock.Lock()
			//for deleteKey, _ := range pbft_utils.PBFTAuthPrepareMap {
			//	splitResult := strings.Split(deleteKey, "_")
			//	if len(splitResult) > 1 {
			//
			//		if splitResult[2] == blockReceive.Head.Hash {
			//			delete(pbft_utils.PBFTAuthPrepareMap, deleteKey)
			//			break
			//		}
			//	}
			//}
			//delete(pbft_utils.PBFTAuthPrepareMap, blockReceive.Head.Hash)
			//pbft_utils.PBFTAuthPrepareMapRWLock.Unlock()
			//
			//utils.Log.Debug("删除 PBFT Commit Auth Map...")
			//pbft_utils.PBFTAuthCommitMapRWLock.Lock()
			//for deleteKey, _ := range pbft_utils.PBFTAuthCommitMap {
			//	splitResult := strings.Split(deleteKey, "_")
			//	if len(splitResult) > 1 {
			//
			//		if splitResult[2] == blockReceive.Head.Hash {
			//			delete(pbft_utils.PBFTAuthCommitMap, deleteKey)
			//			break
			//		}
			//	}
			//}
			//delete(pbft_utils.PBFTAuthCommitMap, blockReceive.Head.Hash)
			//pbft_utils.PBFTAuthCommitMapRWLock.Unlock()

			//insert to local database
			db_api.Db_insertBlockJson(blockReceive.Head.Height, blockReceive.Head.Hash, blockJson)

			GChainMemRWlock.Lock()
			GChainMem.Height = blockReceive.Head.Height
			GChainMem.PrevBlockHash = GChainMem.Hash
			GChainMem.Hash = blockReceive.Head.Hash
			GChainMem.Timestamp = blockReceive.Head.Timestamp
			GChainMemRWlock.Unlock()

			utils.Log.Debug("准备执行区块中的交易！")
			//收到了一个有效的区块，要将区块中的交易执行一遍，并且从交易池中移除, 这两个逻辑在下面这个函数中
			GTransactionPool.CommitTransactionAndExecuteInBlock(blockReceive.TransactionMap, blockReceive.Head.Hash)
			// utils.Log.Info("block's transactions: ", blockReceive)
		} else {
			utils.Log.Info("receive a block json not valid, height:", blockReceive.Head.Height, " Hash:",
				blockReceive.Head.Hash, "prehash:", blockReceive.Head.PrevBlockHash)
		}
	} else {
		utils.Log.Infof("receive block json not valid!  %s", err.Error())
	}

}
func handleSyncBlockJsonMessage(msg *Message, conn Conns) {

	//get the message
	messageByte, _ := json.Marshal(msg.Body)
	messageHash, _ := utils.GetHashFromBytes(messageByte)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, messageHash, messageSignature)
	if !isVerify {
		//if verify failed , end the processing
		utils.Log.Error("Handling received block json verify failed signature not true", msg)
		return
	}

	//receive a block from blockchain
	utils.Log.Debug("receive blockJsonMessage from :", conn.conn.RemoteAddr().String())
	blockJson := msg.Body.([]byte)
	utils.Log.Debug("blockJson", string(blockJson))

	//json解析到结构体里面
	blockReceive := Block{}
	err := json.Unmarshal(blockJson, &blockReceive)
	if err == nil {
		//fixme
		if ValidateBlock(&blockReceive) {

			db_api.Db_insertBlockJson(blockReceive.Head.Height, blockReceive.Head.Hash, blockJson)

			GChainMemRWlock.Lock()
			GChainMem.Height = blockReceive.Head.Height
			GChainMem.PrevBlockHash = GChainMem.Hash
			GChainMem.Hash = blockReceive.Head.Hash
			GChainMem.Timestamp = blockReceive.Head.Timestamp
			GChainMemRWlock.Unlock()

			//收到了一个有效的区块，要将区块中的交易执行一遍，并且从交易池中移除, 这两个逻辑在下面这个函数中
			//block.GTransactionPool.CommitTransactionAndExecuteInBlock(blockReceive.TransactionMap)
			//util.Log( "block's transactions: ", blockReceive)
		} else {
			utils.Log.Error("receive a block json not valid, height:",
				blockReceive.Head.Height, " Hash:", blockReceive.Head.Hash, "prehash:", blockReceive.Head.PrevBlockHash)
		}
	} else {
		utils.Log.Errorf("Handling BlockJsonMessage :  receive block json not valid!  %s", err.Error())
	}

}
// when the local block height has been higher than the chain , require the whole block chain
func handleBlockJsonRenovateMessage(msg *Message, conn Conns) {

	//get the message
	messageByte, _ := json.Marshal(msg.Body)
	messageHash, _ := utils.GetHashFromBytes(messageByte)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, messageHash, messageSignature)
	if !isVerify {
		//if verify failed , end the processing
		utils.Log.Debug("Handling received block json verify failed signature not true", msg)
		return
	}

	//receive a block from blockchain
	utils.Log.Debug("receive blockJsonMessage from :", conn.conn.RemoteAddr().String())
	blockJson := msg.Body.([]byte)
	utils.Log.Debug("blockJson", string(blockJson))

	//json解析到结构体里面
	blockReceive := Block{}
	err := json.Unmarshal(blockJson, &blockReceive)
	if err == nil {
		if ValidateRenovateBlock(&blockReceive) {

			db_api.Db_insertBlockJson(blockReceive.Head.Height,  blockReceive.Head.Hash,blockJson)

			GChainMemRWlock.Lock()
			GChainMem.Height = blockReceive.Head.Height
			GChainMem.PrevBlockHash = GChainMem.Hash
			GChainMem.Hash = blockReceive.Head.Hash
			GChainMem.Timestamp = blockReceive.Head.Timestamp
			GChainMemRWlock.Unlock()

			//收到了一个有效的区块，要将区块中的交易执行一遍，并且从交易池中移除, 这两个逻辑在下面这个函数中
			GTransactionPool.CommitTransactionAndExecuteInBlock(blockReceive.TransactionMap, blockReceive.Head.Hash)
			// utils.Log.Debug("block's transactions: ", blockReceive)
		} else {
			utils.Log.Debug("Handling BlockJsonMessage : receive a block json not valid, height:",
				blockReceive.Head.Height, " Hash:", blockReceive.Head.Hash, "prehash:", blockReceive.Head.PrevBlockHash)
		}
	} else {
		utils.Log.Errorf("Handling BlockJsonMessage :  receive block json not valid!  %s", err.Error())
	}

}
//从其他节点接收到握手信息，处理区块
func handleHandshakeMessage(msg *Message, conn Conns) {
	//要将自己当前的连接情况给对方， 还要将区块全部同步给它们
	handshakeMessage := msg.Body.(HandMessage)
	RemoteHeight := handshakeMessage.Height
	P2pPort := handshakeMessage.P2pPort
	//account address -- we define it on our own
	RemoteAddress := handshakeMessage.Address
	RemoteAccount := handshakeMessage.Account

	utils.Log.Error("Handle Handshake Message:", RemoteAccount)

	//=====================================================================================================================
	//get the local node list and give them the node who wanted to connect to the chain with the seed node naming local
	NodeAccounts := make(map[string]account.AccountInDb)
	NodeAddresses, _ := db_api.Db_getAllAddressAsSlice()
	var NodeAccount account.AccountInDb
	for _, address := range NodeAddresses {
		NodeAccount, _ = account.GetAccountAsStructureFromDB(address)
		NodeAccounts[address] = NodeAccount
	}
	// series the account from sender
	JsonRemoteAccount, err := json.Marshal(RemoteAccount)
	if err != nil {
		utils.Log.Errorf("AccountInfo serilized failed: %s", err.Error())
	}
	tempAccount, err := account.GetAccountAsStructureFromDB(RemoteAddress)
	if err != nil {
		//if there is no this account connecting stored in local
		_ = db_api.Db_insertAccount(JsonRemoteAccount, RemoteAddress)
		_ = db_api.Db_updateAllAddress(RemoteAddress)
	}
	tempAccount, err = account.GetAccountAsStructureFromDB(RemoteAddress)
	utils.Log.Error("Handle Handshake : Check Account Existence Second time! :", err, tempAccount.PublicKey)

	GConnsIpServerPort[conn.conn.RemoteAddr().String()] = P2pPort
	utils.Log.Info("Handling HandshakeMessage:  receive handshake from :",
		conn.conn.RemoteAddr().String(), "remote P2P port : ", P2pPort)
	utils.Log.Info("Handling HandshakeMessage:  remote chain Height:", RemoteHeight, P2pPort)

	var NodeIPs []utils.NODEIP
	//要将自己当前的连接情况给对方，打包进nodeIPS
	GConnsRWlock.RLock()
	for _, connInGConns := range GConns {
		if connInGConns == (Conns{}) {
			continue
		}
		if connInGConns != conn {
			//给非申请节点发送此消息
			remoteIp := connInGConns.conn.RemoteAddr().String() //return like 192.168.0.1:25
			//split the ip and host port
			Index := strings.Index(connInGConns.conn.RemoteAddr().String(), ":")
			ip := remoteIp[0:Index+1] + GConnsIpServerPort[connInGConns.conn.RemoteAddr().String()]
			utils.Log.Info("Handling HandshakeMessage:  packing nodeIP:", ip)
			NodeIPs = append(NodeIPs, utils.NODEIP{IP: ip, RemoteAddress: connInGConns.RemoteAddress})
		}
	}
	GConnsRWlock.RUnlock()
	//目前的前提是连接的节点不会太多，所以考虑一次性发送 , 需要记录每个节点的p2p server port
	err = SendMessage(NodeIpMessage(NodeIPs, NodeAccounts), conn.PtrEncoder)
	if err != nil {
		utils.Log.Error("send nodeIPMessage failed:", conn.conn.RemoteAddr())
	}
	//=====================================================================================================================

	//=====================================================================================================================
	//把当前的全部区块同步给远端，由于两点区块高度差距可能太大，不要一次发送，每个区块发送一次，后续可以考虑5个区块发送一次
	GChainMemRWlock.RLock()
	LocalHeight := GChainMem.Height
	GChainMemRWlock.RUnlock()

	if RemoteHeight == LocalHeight {
		//两方高度一致，什么都不做
	} else if RemoteHeight < LocalHeight {
		//远端高度低，我是种子节点，把本地区块推送过去
		for i := RemoteHeight + 1; i <= LocalHeight; i++ {
			blockData, err := db_api.Db_getBlock(i)
			if err != nil {
				utils.Log.Error(err.Error())
			} else {
				utils.Log.Debug("send:", string(blockData))
				err = SendMessage(SyncBlockJsonMessage(blockData), conn.PtrEncoder)
				if err != nil {
					utils.Log.Error("send blockJson message failed to:", conn.conn.RemoteAddr().String(), "Height:", i)
				}
			}
		}
	} else if RemoteHeight > LocalHeight {
		//本地区块高度低，但是我是种子节点，对面要以我为准。
		for i := int64(1); i <= LocalHeight; i++ {
			blockData, err := db_api.Db_getBlock(i)
			if err != nil {
				utils.Log.Error(err.Error())
			} else {
				utils.Log.Debug("send:", string(blockData))
				err = SendMessage(BlockJsonRenovateMessage(blockData), conn.PtrEncoder)
				if err != nil {
					utils.Log.Error("send blockJson message failed to:", conn.conn.RemoteAddr().String(), "Height:", i)
				}
			}
		}
	}
	//=====================================================================================================================

	tempAccount, err = account.GetAccountAsStructureFromDB(RemoteAddress)
	utils.Log.Error("Check Account Existence Second time! :", err, tempAccount.PublicKey)
}
func handleInitMessage(msg *Message, conn Conns) {
	utils.Log.Debug("receive initMessage from :", conn.conn.RemoteAddr())
}
//我是新节点，我握手完成之后，得到我的种子节点给我返回了消息
//开始处理！！！！！！！！！！！
func handleNodeIpMessage(msg *Message, conn Conns) {

	utils.Log.Debug("receive Node ip from :", conn.conn.RemoteAddr())
	NodeIPs := msg.Body.(NodeIpAndAccountMessage).NodeIPs
	NodeAccounts := msg.Body.(NodeIpAndAccountMessage).NodeAccounts

	//GAddressConnsRWLock.Lock()
	//GAddressConns[msg.Body.(block.NodeIpAndAccountMessage).Address] = conn
	//GAddressConnsRWLock.Unlock()

	var NodeIpMap = make(map[string]bool)

	for _, ip := range NodeIPs {

		if NodeIpMap[ip.IP] {
			continue
		}
		NodeIpMap[ip.IP] = true

		utils.Log.Info("connecting to", ip, "local:", "127.0.0.1:"+GPort)

		conn, err := net.DialTimeout("tcp", ip.IP, time.Second*2)

		if err != nil {
			utils.Log.Error("HandleNodeIpMessage Dial Failed", ip, err.Error())
			continue
		}

		newConn := Conns{conn: conn, PtrEncoder: gob.NewEncoder(conn), RemoteAddress: ip.RemoteAddress}
		utils.Log.Info("HandleNodeIpConnectTo", ip, "Actually", newConn.conn.RemoteAddr().String())
		go HandleConnection(newConn, gob.NewDecoder(conn))

		GConnsRWlock.Lock()
		GConns = append(GConns, newConn)
		GConnsRWlock.Unlock()

		//GAddressConnsRWLock.Lock()
		//GAddressConns[ip.RemoteAddress] = newConn
		//GAddressConnsRWLock.Unlock()

		Index := strings.Index(newConn.conn.RemoteAddr().String(), ":")
		port := newConn.conn.RemoteAddr().String()[Index+1:]
		GConnsIpServerPort[conn.RemoteAddr().String()] = port

	}
	for address, NodeAccount := range NodeAccounts {
		JsonNodeAccount, err := json.Marshal(NodeAccount)
		if err != nil {
			utils.Log.Errorf("AccountInfo serilized failed: %s", err.Error())
		}
		err = db_api.Db_insertAccount(JsonNodeAccount, address)
		if err != nil {
			utils.Log.Errorf("Handling NodeIpMessage insert account into DB error! %s", err.Error())
		} else {
			err = db_api.Db_updateAllAddress(address)
			if err != nil {
				utils.Log.Errorf("Handling NodeIpMessage insert address into DB error! %s", err.Error())
			}
		}

		if address == GLocalIP {
			account.GAccount.IsSubsidized = NodeAccount.IsSubsidized
			account.GAccount.Balance = NodeAccount.Balance
			account.GAccount.Pledge = NodeAccount.Pledge
			account.GAccount.Votes = NodeAccount.Votes
			utils.Log.Info("Handling NodeIpMessage Update local self's account ", account.GAccount)
		}

	}
}
func handleConnectWithAccountMessage(msg *Message, conn Conns) {
	if msg.Body == nil {
		utils.Log.Debug("Handling ConnectMessage failed", msg)
		return
	}
	body := msg.Body.(ConnectMessage)
	err := db_api.Db_updateAllAddress(body.Address)
	if err != nil {
		utils.Log.Errorf("Handling ConnectMessage : updateAllAddress failed %s", err.Error())
		return
	}
	JsonSelfAccount, err := json.Marshal(body.SelfAccount)
	if err != nil {
		utils.Log.Errorf("AccountInfo serilized failed: %s", err.Error())
	}
	_, err = account.GetAccountAsStructureFromDB(body.Address)
	if err != nil {
		_ = db_api.Db_insertAccount(JsonSelfAccount, body.Address)
		_ = db_api.Db_updateAllAddress(body.Address)
	}
	GConnsIpServerPort[conn.conn.RemoteAddr().String()] = body.Port
	return
}
// 收到广播来的交易消息，加入到本地交易池
func handleTransactionMessage(msg *Message, conn Conns) {
	//bMsg, _ := json.MarshalIndent(msg,"","   ")
	//utils.Log.Debug(string(bMsg))

	if msg.Body == nil {
		utils.Log.Errorf("Received tx's body is nil.")
		return
	}
	// 验证消息来源
	messageByte, _ := json.Marshal(msg.Body)
	messageHash, _ := utils.GetHashFromBytes(messageByte)
	isVerify := account.EccVerifyFromHashedData(msg.PublicKey, messageHash, msg.Signature)
	if !isVerify {
		utils.Log.Errorf("Receive tx, but vrf is not ok.")
		return
	}
	transaction := msg.Body.(Transaction)
	// 直接把交易放到交易池添加方法中，具体的处理在底层方法中进行
	GTransactionPool.AddTransaction(transaction)
}
/*
* [From]:               main node
* [Getting&Processing]: replica node
* [Function]:           check the pbft pre-prepare step and broadcast
* [ input  ]:           block.message, tcp connections
 */
func handlePbftPrePrepareMessage(msg *Message, conn Conns) {
	//fix-me 后续的bp之间的连接单独产生，不会复用，所以这里的判断应该没有必要
	//fix-me 现在是全连接，还是有必要的。
	//========judge if i am in this period to produce block========
	//need to be the producer, every replicas should broadcast the prepare message


	utils.Log.Debug(" 收到Pre prepare消息. ")

	var CanProduce = false
	GChainMemRWlock.RLock()
	for _, str := range GChainMem.ProduceAccount {
		if str == GMyAccountAddress {
			CanProduce = true
			break
		}
	}
	GChainMemRWlock.RUnlock()

	if !CanProduce {
		//utils.Log.Debug("Handling PbftPrePrepareMessage :	I am not producer:", block.GMyAccountAddress)
		return
	}

	utils.Log.Debug("进行 Pre Prepare 消息验证.")
	//================================================
	//utils.Log.Debug("Handling PbftPrePrepareMessage :	receive pre-prepare Message from :", conn.conn.RemoteAddr())

	//============verify the signature=============
	blockReceive := msg.Body.(Block)
	blockHash := []byte(blockReceive.Head.Hash)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, blockHash, messageSignature)
	//======================================

	pubKeyPeer := utils.Base58Encode(msg.PublicKey)
	keyVerify := strconv.FormatInt(blockReceive.Head.Height, 16) + "_" + string(pubKeyPeer) + "_" + blockReceive.Head.Hash

	if isVerify {

		CONSENSUS_MAP_LOCK.Lock()
		pass, ok := CONSENSUS_MAP[blockReceive.Head.Hash]
		if !ok {
			CONSENSUS_MAP[blockReceive.Head.Hash] = false
		} else {
			if pass == true {
				return
			}
		}
		CONSENSUS_MAP_LOCK.Unlock()

		pbft_utils.PBFTAuthPrePrepareMapRWLock.Lock()
		signatureThis, key := pbft_utils.PBFTAuthPrePrepareMap[keyVerify]
		if key {
			if bytes.Equal(signatureThis, pbft_utils.PBFTAuthPrePrepareMap[keyVerify]) {
				//接收到的和之前接受的到pre-prepare中的签名相同
				//fixmed : if multiple get , how to do ?
				//P2pBroadcastPrepare(blockReceive)
			} else {
				utils.Log.Debug("Handling Pbft pre-PrepareMessage failed illegal!! signature not the same :")
			}
		} else {

			//todo : when to delete the key-value from this map
			isLegal := false
			GChainMemRWlock.RLock()
			if GChainMem.Hash == blockReceive.Head.PrevBlockHash &&
				GChainMem.Height+1 == blockReceive.Head.Height {
				//confirm this pre-prepare message has been received
				pbft_utils.PBFTAuthPrePrepareMap[keyVerify] = msg.Signature
				pbft_utils.PBFTAuthPrePrepareMap[blockReceive.Head.Hash] = []byte("exist")
				isLegal = true
			}
			GChainMemRWlock.RUnlock()
			if isLegal {
				P2pBroadcastPrepare(blockReceive)
			} else {
				utils.Log.Error("Handling Pbft pre-PrepareMessage failed illegal!! transaction length is zero:")
			}
		}
		pbft_utils.PBFTAuthPrePrepareMapRWLock.Unlock()

	}

}
/*
* [From]:               replica node
* [Getting&Processing]: all super node
* [Function]:           check the pbft prepare step and sign
* [ input  ]:           block.message, tcp connections
 */
func handlePbftPrepareMessage(msg *Message, conn Conns) {
	//utils.Log.Debug("receive prepare Message from :", conn.conn.RemoteAddr())
	//========judge if i am in this period to produce block========
	//need to be the producer, every replicas should broadcast the prepare message
	var CanProduce = false
	GChainMemRWlock.RLock()
	for _, str := range GChainMem.ProduceAccount {
		if str == GMyAccountAddress {
			CanProduce = true
			break
		}
	}
	GChainMemRWlock.RUnlock()

	if !CanProduce {
		//utils.Log.Info("I am not producer:", block.GMyAccountAddress)
		return
	}
	//================================================

	//============verify the signature=============
	blockReceive := msg.Body.(Block)
	blockHash := []byte(blockReceive.Head.Hash)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, blockHash, messageSignature)
	//======================================
	var isBeginToAuth = false
	pubKeyPeer := utils.Base58Encode(msg.PublicKey)
	keyVerify := strconv.FormatInt(blockReceive.Head.Height, 16) + "_" + string(pubKeyPeer) + "_" + blockReceive.Head.Hash

	if isVerify {
		//验证这个请求是否合法
		pbft_utils.PBFTAuthPrepareMapRWLock.Lock()
		signatureThis, key := pbft_utils.PBFTAuthPrepareMap[keyVerify]
		if key {
			if bytes.Equal(signatureThis, pbft_utils.PBFTAuthPrepareMap[keyVerify]) {
				//接收到的和之前接受的到prepare相同
				utils.Log.Debug("Handling Pbft Prepare Message===========multi============")
				//不进行计数
				//fixme : if multiple get , how to do ?
				//return
			} else {
				//判断为非法
				utils.Log.Debug("Handling Pbft Prepare Message failed illegal!! signature not the same :")
				return
			}
		} else {
			//to-do : when to delete the key-value from this map
			//第一次收到，开始验证
			utils.Log.Debug("Handling Pbft Prepare  message ============first ===========")

			pbft_utils.PBFTAuthPrePrepareMapRWLock.RLock()
			//检查是否收到过这个block的pre-prepare消息
			_, isExist := pbft_utils.PBFTAuthPrePrepareMap[blockReceive.Head.Hash]
			pbft_utils.PBFTAuthPrePrepareMapRWLock.RUnlock()
			//检查其他是否合法,记得枷锁！！！！
			GChainMemRWlock.RLock()
			if GChainMem.Hash == blockReceive.Head.PrevBlockHash &&
				GChainMem.Height+1 == blockReceive.Head.Height &&
				isExist {

				pbft_utils.PBFTAuthPrepareMap[keyVerify] = msg.Signature
				isBeginToAuth = true
			}
			GChainMemRWlock.RUnlock()

		}
		pbft_utils.PBFTAuthPrepareMapRWLock.Unlock()
	}

	if isBeginToAuth {

		//获取参与共识的节点的个数，pbft过程需要
		//需要收到2f以上的节点的prepare才进入认证。
		GChainMemRWlock.RLock()
		ProducerNodeNum := len(GChainMem.ProduceAccount)
		GChainMemRWlock.RUnlock()

		pbft_utils.PBFTVoteAuthPrepareMapLock.Lock()
		//进入到这里，说明请求合法，可以直接增加票数了
		num, ok := pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Head.Hash]
		if !ok {
			//pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Height] = 1
			pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Head.Hash] = 1
		}
		pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Head.Hash] = num + 1

		if pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Head.Hash] > (ProducerNodeNum*2)/3-2 {
		//if pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Head.Hash] >= 1 {

			delete(pbft_utils.PBFTVoteAuthPrepareMap, blockReceive.Head.Hash)
			P2pBroadcastCommit(blockReceive)
		} else {
			//util.Log("prepare vote counting result ：")
			//util.Log(pbft_utils.PBFTVoteAuthPrepareMap[blockReceive.Hash])
			//util.Log("prepare vote counting result should larger than：")
			//util.Log((ProducerNodeNum / 3) * 2)
			utils.Log.Error("Handling Pbft Prepare fail count not enough")

		}
		pbft_utils.PBFTVoteAuthPrepareMapLock.Unlock()
	}

}
/*
* [From]:               all super node
* [Getting&Processing]: all super node
* [Function]:           check the commit step
* [ input  ]:           block.message, tcp connections
 */
func handlePbftCommitMessage(msg *Message, conn Conns) {
	//utils.Log.Info("Handling commit message from", conn.conn.RemoteAddr())
	//========judge if i am in this period to produce block========
	//need to be the producer, every replicas should broadcast the prepare message
	var CanProduce = false
	GChainMemRWlock.RLock()
	for _, str := range GChainMem.ProduceAccount {
		if str == GMyAccountAddress {
			CanProduce = true
			break
		}
	}
	GChainMemRWlock.RUnlock()

	if !CanProduce {
		//utils.Log.Debug("I am not producer:", block.GMyAccountAddress)
		return
	}

	//============verify the signature=============
	blockReceive := msg.Body.(Block)
	blockHash := []byte(blockReceive.Head.Hash)
	//get the signature
	messageSignature := msg.Signature
	messagePubKeyByte := msg.PublicKey
	//verify the context and signature
	isVerify := account.EccVerifyFromHashedData(messagePubKeyByte, blockHash, messageSignature)
	//util.Log(blockReceive)

	pubKeyPeer := utils.Base58Encode(msg.PublicKey)
	keyVerify := strconv.FormatInt(blockReceive.Head.Height, 16) + "_" + string(pubKeyPeer) + "_" + blockReceive.Head.Hash
	var isBeginToAuth = false
	var isPassConsensus = false

	if isVerify {

		//验证这个请求是否合法
		pbft_utils.PBFTAuthCommitMapRWLock.Lock()
		signatureThis, key := pbft_utils.PBFTAuthCommitMap[keyVerify]
		if key {
			if bytes.Equal(signatureThis, pbft_utils.PBFTAuthCommitMap[keyVerify]) {
				//接收到的和之前接受的到prepare相同
				//不进行计数
				utils.Log.Debug("Handling Pbft commit Message successfully =======multiple==========:")
			} else {
				//判断为非法
				utils.Log.Debug("Handling Pbft commit Message failed illegal!! signature not the same :")
				return
			}
		} else {
			//第一次收到，开始验证
			//to-do : when to delete the key-value from this map
			utils.Log.Debug("Handling Pbft commit Message successfully =======first==========:")
			pbft_utils.PBFTAuthCommitMap[keyVerify] = msg.Signature
			isBeginToAuth = true

			pbft_utils.PBFTAuthPrePrepareMapRWLock.RLock()
			//检查是否收到过这个block的pre-prepare消息
			_, isExist := pbft_utils.PBFTAuthPrePrepareMap[blockReceive.Head.Hash]
			pbft_utils.PBFTAuthPrePrepareMapRWLock.RUnlock()

			GChainMemRWlock.RLock()
			if GChainMem.Hash == blockReceive.Head.PrevBlockHash &&
				GChainMem.Height+1 == blockReceive.Head.Height &&
				isExist {
				pbft_utils.PBFTAuthCommitMap[keyVerify] = msg.Signature
				isBeginToAuth = true

			}
			GChainMemRWlock.RUnlock()
		}
		pbft_utils.PBFTAuthCommitMapRWLock.Unlock()

	}

	if isBeginToAuth {

		GChainMemRWlock.RLock()
		ProducerNodeNum := len(GChainMem.ProduceAccount)
		GChainMemRWlock.RUnlock()

		//开始计算收到的commit个数
		pbft_utils.PBFTVoteAuthCommitMapLock.Lock()
		num, ok := pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash]
		if !ok {
			pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash] = 1
		}
		pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash] = num + 1

		if pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash] > (ProducerNodeNum*2)/3-1 {
		//if pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash] >= 1 {
			utils.Log.Critical("PBFT 通过，票数：",pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash])
			//prepare 的 2n/3的容错性
			//进入final 阶段！！ 区块确定了，出！
			utils.Log.Debug(" block commit pbft success counting enough -- now ")

			delete(pbft_utils.PBFTVoteAuthCommitMap, blockReceive.Head.Hash)

			isPassConsensus = true

		} else {
			//util.Log("commit vote counting result ：")
			//util.Log(pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Hash])
			//util.Log("commit vote counting result should larger than：")
			//util.Log((ProducerNodeNum / 3) * 2)
			utils.Log.Debug(" Commit votes : ",pbft_utils.PBFTVoteAuthCommitMap[blockReceive.Head.Hash])
			//共识没有通过，丢掉
		}
		pbft_utils.PBFTVoteAuthCommitMapLock.Unlock()

		//获取参与共识的节点的个数，pbft过程需要
		//需要收到2f以上的节点的commit才进入认证。
		if isPassConsensus {
			//共识通过了
			blockJson, _ := json.Marshal(blockReceive)
			GChainMemRWlock.Lock()
			//如果自己是生产者。验证通过，并且加入主链，再传播给所有节点
			if  GMyAccountAddress == blockReceive.Head.Producer &&
				GChainMem.Height+1 == blockReceive.Head.Height &&
				GChainMem.Hash == blockReceive.Head.PrevBlockHash {

				utils.Log.Debug("block transaction info :", blockReceive.Head)
				db_api.Db_insertBlockJson(blockReceive.Head.Height, blockReceive.Head.Hash, blockJson)

				//把新的区块广播给所有节点
				P2pBroadcastToAllNodesBlock(blockJson)

				GChainMem.Height = blockReceive.Head.Height
				GChainMem.PrevBlockHash = GChainMem.Hash
				GChainMem.Hash = blockReceive.Head.Hash
				GChainMem.Timestamp = blockReceive.Head.Timestamp

				//执行区块中交易的操作到下面这个函数去改
				//这是当自己是出块节点,同时成功出块了,对自己执行,因为广播的时候不会广播给自己
				execBlock := Block{}
				json.Unmarshal(blockJson,&execBlock)
				GTransactionPool.CommitTransactionAndExecuteInBlock(execBlock.TransactionMap, execBlock.Head.Hash)
			}
			GChainMemRWlock.Unlock()

			//delete the useless in map
			pbft_utils.PBFTAuthPrePrepareMapRWLock.Lock()
			delete(pbft_utils.PBFTAuthPrePrepareMap, keyVerify)
			delete(pbft_utils.PBFTAuthPrePrepareMap, blockReceive.Head.Hash)
			pbft_utils.PBFTAuthPrePrepareMapRWLock.Unlock()

			pbft_utils.PBFTAuthPrepareMapRWLock.Lock()
			delete(pbft_utils.PBFTAuthPrepareMap, keyVerify)
			delete(pbft_utils.PBFTAuthPrepareMap, blockReceive.Head.Hash)
			pbft_utils.PBFTAuthPrepareMapRWLock.Unlock()

			pbft_utils.PBFTAuthCommitMapRWLock.Lock()
			delete(pbft_utils.PBFTAuthCommitMap, keyVerify)
			delete(pbft_utils.PBFTAuthCommitMap, blockReceive.Head.Hash)
			pbft_utils.PBFTAuthCommitMapRWLock.Unlock()

		}
	}
}

/**********************************************************************************************************************
										       P2P 建立连接过程处理函数
**********************************************************************************************************************/
func P2pHandleConnection(conn Conns, dec *gob.Decoder) {
	//END_PEER_CONNECTION:
	for {
		var msg Message
		//客户端主动关闭连接会产生EOF，不对此处理会导致goroutine空转,退出协程
		err := receiveMessage(&msg, dec)
		if io.EOF == err {
			utils.Log.Error("EOF:", conn.RemoteAddress)
			return
		} else if err != nil {
			utils.Log.Error(conn.RemoteAddress, ",error:", err.Error())
			continue
		} else {
			ProcessMessage(&msg, conn)
		}
	}
}
func receiveMessage(msg *Message, dec *gob.Decoder) error {
	err := dec.Decode(msg)
	//utils.Log.Debug("Receive msg type: ", msg.Type)

	if err != nil {
		if err == io.EOF {
			return err
		}
		log.Println("[Receive Message] %s", err.Error())
	}
	return err
}
func ConnectSeedNodes(p2pPeerAddress []string, Height int64, p2pListenPort string, selfAccount account.Account, address string) bool {
	count := 0

	for _, peerAddr := range p2pPeerAddress {
		conn, err := Connect(peerAddr, Height, p2pListenPort)
		if err != nil {
			utils.Log.Error(err.Error())
			count++
		} else {
			//fix-me conn似乎有必要保存作为长连接，EOS中的长连接作用尚不清楚
			//util.Log("[",utils.GetUTCTime(),"]  ",conn)
			enc := gob.NewEncoder(conn)
			newConns := Conns{conn: conn, PtrEncoder: enc}

			accountToSend := account.TransferToAccountInDb(selfAccount)
			//发送握手信息主要是为了尽量区块高度一致
			err := SendMessage(HandshakeMessage(Height, p2pListenPort, accountToSend, address), enc)
			if err != nil {
				utils.Log.Errorf("Connecting to SeedNodes  :  send handshake failed %s", err.Error())
			}

			//SendMessage(InitMessage(gob.NewDecoder(conn))
			Index := strings.Index(newConns.conn.RemoteAddr().String(), ":")
			port := newConns.conn.RemoteAddr().String()[Index+1:]

			GConnsRWlock.Lock()
			GConns = append(GConns, newConns)
			GConnsIpServerPort[conn.RemoteAddr().String()] = port
			GConnsRWlock.Unlock()
			//处理异步消息
			//go HandleConnection(newConns, gob.NewDecoder(conn))
			go P2pHandleConnection(newConns, gob.NewDecoder(conn))
		}

	}
	//util.Log( "Peer count", len(p2p_peer_address))

	if count == len(p2pPeerAddress) && count != 0 {
		return false
	} else {
		return true
	}
}
