package Network
import (
	"DAG-Exp/src/block"
	"DAG-Exp/src/dag"
	"DAG-Exp/src/sharecc"
	"DAG-Exp/src/utils"
	"encoding/gob"
	"log"
	"sync"
	"syscall"
)

var sendMessageMutex sync.Mutex



// 消息发送出口
func SendMessage(msg *block.Message, enc *gob.Encoder) error {
	// utils.Log.Debug(msg)
	sendMessageMutex.Lock()
	err := enc.Encode(msg)
	sendMessageMutex.Unlock()
	if err != nil {
		log.Println("[Send Message] %s", err.Error())
	}
	return err
}
func BroadcastSyncTestMsg(msg sharecc.SyncTest) {
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		//utils.Log.Debug("Broadcast msg to:")
		//utils.Log.Debug(conn)
		if conn == (Conns{}) {
			utils.Log.Errorf("BroadcastSyncTestMsg(msg sharecc.SyncTest) 29.")
			// utils.Log.Debug(conn)
			continue
		}
		err := SendMessage(block.SyncMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2pBroadcastToSuperNodeTransaction send transaction failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}

/*
	广播 DAG Tag 消息

*/
func BroadcastDagTag(msg dag.DagTag){
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		//utils.Log.Debug("Broadcast msg to:", conn)
		if conn == (Conns{}) {
			// utils.Log.Debug(conn)
			continue
		}
		err := SendMessage(block.DagTagMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2p Broadcast Send Transaction Failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}

func BroadcastEpoch(msg dag.EpochTag){
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		//utils.Log.Debug("Broadcast msg to:", conn)
		if conn == (Conns{}) {
			// utils.Log.Debug(conn)
			continue
		}
		err := SendMessage(block.EpochMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2p Broadcast Send Transaction Failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}

func BroadcastVote(msg dag.Vote){
	GConnsRWlock.RLock()
	for index, conn := range GConns {
		//utils.Log.Debug("Broadcast msg to:", conn)
		if conn == (Conns{}) {
			// utils.Log.Debug(conn)
			continue
		}
		err := SendMessage(block.VoteMessage(msg), conn.PtrEncoder)
		if err == syscall.EPIPE {
			utils.Log.Errorf("P2p Broadcast Send Transaction Failed,  %s", conn.conn.RemoteAddr().String())
			GConns[index] = Conns{}
			delete(GConnsIpServerPort, conn.conn.RemoteAddr().String())
		}
	}
	GConnsRWlock.RUnlock()
}