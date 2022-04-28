package Network

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/block"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/db_api"
	"DAG-Exp/src/sharecc"
	"DAG-Exp/src/utils"
	"encoding/gob"
	"encoding/json"
	"net"
	"strings"
	"time"
)

func ProcessMessage(msg *block.Message, conn Conns) {
	utils.Log.Debug("处理消息.....", msg.Type)
	switch msg.Type {
	// 系统启动时需要的消息处理方法
	case block.MessageTypeInit:
		handleInitMessage(msg, conn)
	case block.MessageTypeHandshake:
		handleHandshakeMessage(msg, conn)
	case block.MessageTypeNodeIp:
		handleNodeIpMessage(msg, conn)
	case block.MessageTypeConnect:
		handleConnectWithAccountMessage(msg, conn)
	// 应用层逻辑方法
	case block.MessageTypeSyncTest:
		handleSyncTest(msg)
	// Top Tag 消息
	case block.MessageTypeCreateDag:
		handleCreateDag(msg)
	// Request Tag 消息
	case block.MessageTypeRequestData:
		handleRequestData(msg)
	// Response Tag 消息
	case block.MessageTypeResponse:
		handleResponseData(msg)
	// Ack Tag 消息
	case block.MessageTypeAck:
		handleAck(msg)
	case block.MessageSubmit:
		handleSubmit(msg)
	case block.MessageVote:
		handleVote(msg)
	default:
	}
}
/* 系统层方法 */
//从其他节点接收到握手信息，处理区块
func handleHandshakeMessage(msg *block.Message, conn Conns) {
	handshakeMessage := msg.Body.(block.HandMessage)
	P2pPort := handshakeMessage.P2pPort
	//account address -- we define it on our own
	RemoteAddress := handshakeMessage.Address
	RemoteAccount := handshakeMessage.Account

	utils.Log.Debug("Handle Handshake Message:", RemoteAddress)

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
	// utils.Log.Error("Handle Handshake : Check Account Existence Second time! :", err, tempAccount.PublicKey)

	GConnsIpServerPort[conn.conn.RemoteAddr().String()] = P2pPort
	utils.Log.Info("Handling HandshakeMessage:  receive handshake from :",
		conn.conn.RemoteAddr().String(), "remote P2P port : ", P2pPort)
	utils.Log.Info("Handling HandshakeMessage: ",  P2pPort)

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
	err = SendMessage(block.NodeIpMessage(NodeIPs, NodeAccounts), conn.PtrEncoder)
	if err != nil {
		utils.Log.Error("send nodeIPMessage failed:", conn.conn.RemoteAddr())
	}


	tempAccount, err = account.GetAccountAsStructureFromDB(RemoteAddress)
	utils.Log.Error("Check Account Existence Second time! :", err, string(tempAccount.PublicKey))
}
func handleInitMessage(msg *block.Message, conn Conns) {
	utils.Log.Debug("receive initMessage from :", conn.conn.RemoteAddr())
}
//我是新节点，我握手完成之后，得到我的种子节点给我返回了消息
func handleNodeIpMessage(msg *block.Message, conn Conns) {

	utils.Log.Debug("receive Node ip from :", conn.conn.RemoteAddr())
	NodeIPs := msg.Body.(block.NodeIpAndAccountMessage).NodeIPs
	NodeAccounts := msg.Body.(block.NodeIpAndAccountMessage).NodeAccounts

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
			account.GAccount.Balance = NodeAccount.Balance
			utils.Log.Info("Handling NodeIpMessage Update local self's account ", account.GAccount)
		}

	}
}
func handleConnectWithAccountMessage(msg *block.Message, conn Conns) {
	if msg.Body == nil {
		utils.Log.Debug("Handling ConnectMessage failed", msg)
		return
	}
	body := msg.Body.(block.ConnectMessage)
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


/* 应用层方法 */
// 测试用
func handleSyncTest(msg *block.Message) {
	utils.Log.Debug("处理收到的测试消息")
	bMsg, _ := json.Marshal(msg.Body)
	var syncTestMsg sharecc.SyncTest
	err := json.Unmarshal(bMsg, &syncTestMsg)
	if err != nil {
		utils.Log.Error("Sync Test Msg 解析失败！")
		return
	}
	utils.Log.Debug(syncTestMsg)
	return
}

func handleCreateDag(msg *block.Message) {
	// 1.1 验证签名
	// 1.2 检查本地区块链数据

	// 2. 验证通过存入本地数据库
	body := msg.Body.(dag.DagTag)
	err := db_api.LevelDbDagSaveTag(body)
	if err != nil {
		utils.Log.Error("Cannot create DAG. Error: ", err)
	}
	return
}

func handleRequestData(msg *block.Message) {
	// 1. 解析 Tag 数据，验证数据
	tag := msg.Body.(dag.DagTag)
	tagPayload := tag.Body.(dag.RequestTagBody)
	utils.Log.Debug("处理 Request Data Tag. 来自：",tagPayload.FromAddress,
		"\n 目标：",tagPayload.ToAddress,
		"\n 自己：",string(account.GAccount.Address))
	// 1.1 获取请求目的地
	// 1.2 如果请求目标是自己，处理数据请求
	if tagPayload.ToAddress == string(account.GAccount.Address) {
		utils.Log.Debug("目标是自己，处理数据请求...")
		// 1.2.1 验证权限
		pass := dag.ProcessRequestData()
		// 1.2.2 组装 Response Tag
		var prevHash []string
		prevHash = append(prevHash, tag.Hash)
		responseTagBody := dag.ResponseTagBody{
			FromAddress: string(account.GAccount.Address),
			ToAddress:   tagPayload.FromAddress,
			ExecResult:  pass,
			InfoURL:     "test - URL",
			Code:        "test - Access Code",
			ReqHash:     tag.Hash,
			PrevHash:    tag.PrevHash,
		}
		responseTag := dag.DagTag{
			TimeStamp: time.Now().UnixNano(),
			Depth:     tag.Depth,
			PrevHash:  prevHash,
			TagType:   dag.RESPONSETAG,
			Body:      responseTagBody,
			Miner:     string(account.GAccount.Address),
			Signature: string(account.GAccount.Address) + "'s Signature",
		}
		responseTag.Hash = dag.CalculateTagHash(responseTag)
		bResponseTag, _ := json.MarshalIndent(responseTag,"","\t")
		utils.Log.Debug("完成处理，生成 Response Tag: ", bResponseTag)
		// 广播 Response
		BroadcastDagTag(responseTag)
		// 存储 Response Tag
		err := db_api.LevelDbDagSaveTag(responseTag)
		if err != nil {
			utils.Log.Error("Save Response Tag Failed. Error:", err.Error())
		}
	}
	// 1.3 不管目标是不是自己，存储 Request tag
	err := db_api.LevelDbDagSaveTag(tag)
	if err != nil {
		utils.Log.Error("Save Request Tag Failed. Error: ", err.Error())
	}
}

func handleResponseData(msg *block.Message) {
	// 1. 解析 Response Tag
	tag := msg.Body.(dag.DagTag)
	tagPayload := tag.Body.(dag.ResponseTagBody)
	err := db_api.LevelDbDagSaveTag(tag)
	// 2. 如果目标是自己
	if tagPayload.ToAddress == string(account.GAccount.Address) {
		// 2.1 检查结果
		isOk := tagPayload.ExecResult
		// 2.2 组装 Ack Tag
		ackTagBody := dag.AckTagBody{
			FromAddress: string(account.GAccount.Address),
			ToAddress:   tagPayload.FromAddress,
			Pass:        isOk,
			ReqHash:     tagPayload.ReqHash,
			RespHash:    tag.Hash,
			PrevHash:    tagPayload.PrevHash,
		}
		var prevHash []string
		prevHash = append(prevHash, tag.Hash)
		ackTag := dag.DagTag{
			TimeStamp: time.Now().UnixNano(),
			Depth:     tag.Depth,
			PrevHash:  prevHash,
			TagType:   dag.ACKTAG,
			Body:      ackTagBody,
			Miner:     string(account.GAccount.Address),
			Signature: string(account.GAccount.Address) + "'s Signature",
		}
		ackTag.Hash = dag.CalculateTagHash(ackTag)
		// 2.3 广播 Ack Tag
		BroadcastDagTag(ackTag)
		// 2.4 存储 Ack Tag
		err := db_api.LevelDbDagSaveTag(ackTag)
		if err != nil {
			utils.Log.Error("Save ACK Tag Failed. Error: ", err.Error())
		}
	}
	// 3. 不管目标是不是自己，验证 Response Tag 并存储
	// 3.1 验证信息可靠性，并存储到数据库
	err = db_api.LevelDbDagSaveTag(tag)
	if err != nil {
		utils.Log.Error("Save Response Tag Failed. Error: ", err.Error())
	}
}

func handleAck(msg *block.Message) {
	// 存储数据
	tag := msg.Body.(dag.DagTag)
	err := db_api.LevelDbDagSaveTag(tag)

	// 计算校验和
	// tagPayload := tag.Body.(dag.AckTagBody)
	// db_api.SaveChecksum(tag.Hash, tagPayload.PrevHash)
	
	if err != nil {
		utils.Log.Error("Save Ack Tag Failed. Error: ", err)
	}
	return
}

func handleSubmit(msg *block.Message) {
	pass := 0
	// 1. 检查签名，哈希
	epochTag := msg.Body.(dag.EpochTag)
	epochHash := epochTag.Hash
	//epochTag.Hash = ""
	//checkHash := dag.CalculateEpochTagHash(epochTag)
	//if checkHash == epochHash {
	//	utils.Log.Debug("哈希校验通过")
	//}
	// 2. 检查本地是否包含对应的tag
	//checksum := ""
	//for _, v := range epochTag.PrevHashList {
	//	ok := db_api.GetTag(v)
	//	if ok == nil {
	//		pass = -1
	//		break
	//	}
	//	checksum = checkHash + v
	//}
	// 3. 检查校验和（用列表里的所有Hash进行一次求和哈希运算模拟）
	//hash,_ := utils.GetHashFromBytes([]byte(checksum))
	//utils.Log.Debug("校验和： ", hash)
	// 4. 通过检查，广播同意票（9个节点，6票通过）
	if pass == 0 {
		utils.Log.Debug("新的Epoch：", epochHash)
		dag.VOTELOCK.Lock()
		utils.Log.Debug("累加投票结果：", epochHash)
		dag.VOTE[epochHash] = dag.VOTE[epochHash] + 1
		dag.VOTELOCK.Unlock()
		vote := dag.Vote{
			EpochHash: epochHash,
			TimeStamp: epochTag.TimeStamp,
			Pass:      0,
			Latency:   0,
		}
		utils.Log.Debug("广播我的投票：", vote)
		BroadcastVote(vote)
	}
}

func handleVote(msg *block.Message){
	vote := msg.Body.(dag.Vote)
	utils.Log.Debug("收到投票：", vote.EpochHash, vote.Pass)
	if vote.Pass == 1 {
		dag.VSLOCK.Lock()
		dag.VoteLatencyResult[vote.EpochHash] = append(dag.VoteLatencyResult[vote.EpochHash], vote.Latency)
		utils.Log.Debug("统计的时延结果：", dag.VoteLatencyResult)
		dag.VSLOCK.Unlock()
		return
	}

	epochHash := vote.EpochHash
	//dag.VOTELOCK.Lock()
	//n := dag.VOTE[epochHash]
	//dag.VOTELOCK.Unlock()
	//if n >= 4 {
	//	return
	//}
	dag.VOTELOCK.Lock()
	dag.VOTE[epochHash] = dag.VOTE[epochHash] + 1
	if dag.VOTE[epochHash] == 4 {
		t := time.Now().UnixNano()
		//latency := (t - vote.TimeStamp)/ int64(time.Microsecond) Millisecond
		latency := (t - vote.TimeStamp)/ int64(time.Millisecond)
		if dag.VoteLatency[epochHash] == 0 {
			dag.VoteLatency[epochHash] = latency
		}
		utils.Log.Debug("投票统计完成：", epochHash, "，时延：", dag.VoteLatency[epochHash])
		voteResult := dag.Vote{
			EpochHash: epochHash,
			TimeStamp: 0,
			Pass:      1,
			Latency:   latency,
		}
		BroadcastVote(voteResult)
		utils.Log.Debug("已经统计到 4 票通过，核验通过，广播完成")
	}
	dag.VOTELOCK.Unlock()
	return
}