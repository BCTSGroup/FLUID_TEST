package block
import (
	"bfc/src/sharecc"
	"bfc/src/utils"
	"encoding/gob"
	"log"
	"sync"
	"syscall"
)

var sendMessageMutex sync.Mutex

// 消息发送出口
func SendMessage(msg *Message, enc *gob.Encoder) error {
	sendMessageMutex.Lock()
	err := enc.Encode(msg)
	sendMessageMutex.Unlock()
	if err != nil {
		log.Println("[Send Message] %s", err.Error())
	}
	return err
}

// 共享数据消息的广播
func P2pBroadcastSubmitSharingDataRequest(msg sharecc.SubmitRequest) {
	// utils.Log.Infof("P2pBroadcastToSuperNodeTransaction, peers:%d", len(GConns))
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(SubmitSharingDataMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2pBroadcastToSuperNodeTransaction send transaction failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}

// 共享数据消息的广播
func P2pBroadcastGetSharedDataRequest(msg sharecc.GetSharedDataRequest) {
	// utils.Log.Infof("P2pBroadcastToSuperNodeTransaction, peers:%d", len(GConns))
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(GetSharedDataRequestMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2pBroadcastToSuperNodeTransaction send transaction failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}

// 原始交易信息广播给其他节点
func P2pBroadcastTransaction(msg Transaction, operatorType int) {
	// utils.Log.Infof("P2pBroadcastToSuperNodeTransaction, peers:%d", len(GConns))
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(BlockTransactionMessage(msg, operatorType), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2pBroadcastToSuperNodeTransaction send transaction failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}
// 新生成的区块广播给所有节点
func P2pBroadcastToAllNodesBlock(newBlock []byte) {

	GConnsRWlock.RLock()
	defer GConnsRWlock.RUnlock()
	utils.Log.Infof("P2pBroadcastToAllNodesBlock  :  Broadcast, peers:%d", len(GConns))

	//GAddressConnsRWLock.RLock()
	//util.Log(len(GAddressConns))
	//util.Log("======================")
	//util.Log(GAddressConns)
	//GAddressConnsRWLock.RUnlock()

	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(BlockJsonMessage(newBlock), conn.PtrEncoder)
		if err != nil {
			utils.Log.Errorf("send block failed,%v", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}

}
// PBFT pre-prepare 消息
func P2pBroadcastPrePrepare(newBlock *Block) {

	GConnsRWlock.RLock()
	defer GConnsRWlock.RUnlock()
	utils.Log.Info("peers:", len(GConns))
	for index, conn := range GConns {
		//fixed　：　当此连接的对面是出块节点的时候
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(PbftPrePrepareMessage(*newBlock), conn.PtrEncoder)
		if err != nil {
			utils.Log.Error("send pbft preprepare failed", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
}
// PBFT prepare 消息
func P2pBroadcastPrepare(newBlock Block) {

	GConnsRWlock.RLock()
	defer GConnsRWlock.RUnlock()
	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(PbftPrepareMessage(newBlock), conn.PtrEncoder)
		if err != nil {
			utils.Log.Error("send block failed", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}

}
// PBFT Commit 消息
func P2pBroadcastCommit(newBlock Block) {
	if err := PreExecuteBlock(&newBlock, false); err != nil {
		utils.Log.Error("error in preExecuteBlock")
		return
	}

	GConnsRWlock.RLock()
	utils.Log.Info("P2pBroadcastCommit  :  peers:", len(GConns))
	for index, conn := range GConns {
		if conn == (Conns{}) {
			continue
		}
		err := SendMessage(PbftCommitMessage(newBlock), conn.PtrEncoder)
		if err != nil {
			utils.Log.Error("send block failed", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}
