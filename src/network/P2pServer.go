package Network

import (
	"DAG-Exp/src/account"
	"DAG-Exp/src/block"
	"DAG-Exp/src/utils"
	"context"
	"encoding/gob"
	"io"
	"log"
	"net"
	"strings"
)

var GPort string //for single host
var GLocalIP string

// 新建 P2P 服务
func NewP2pServer(ctx context.Context, listenPort string) net.Listener {
	// 开启端口监听
	utils.Log.Info("Setting up NewP2pServer  :  p2p listen port:" + GLocalIP + ":" + listenPort)
	listener, err := net.Listen("tcp", GLocalIP+":"+listenPort)
	if err != nil {
		log.Println("Setting up NewP2pServer  :  NewServer Failed")
		return nil
	}
	// 新协程，等待其他节点的加入
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
			GConnsIpServerPort[conn.RemoteAddr().String()] = "None"
			GConnsRWlock.Unlock()

			// 创建connection的解码器，并传到P2pHandleConnection中进行解码
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

// 主动连接种子节点
func ConnectSeedNodes(p2pPeerAddress []string, p2pListenPort string, selfAccount account.Account, address string) bool {
	count := 0

	for _, peerAddr := range p2pPeerAddress {
		conn, err := Connect(peerAddr, p2pListenPort)
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
			err := SendMessage(block.HandshakeMessage(p2pListenPort, accountToSend, address), enc)
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

//
func P2pHandleConnection(conn Conns, dec *gob.Decoder) {
	//END_PEER_CONNECTION:
	for {
		var msg block.Message
		//客户端主动关闭连接会产生EOF，不对此处理会导致goroutine空转,退出协程
		utils.Log.Debugf("接收消息前")
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
// 接收对等节点的消息
func receiveMessage(msg *block.Message, dec *gob.Decoder) error {
	err := dec.Decode(msg)

	if err != nil {
		if err == io.EOF {
			return err
		}
		log.Println("[Receive Message] %s", err.Error())
	}
	return err
}

