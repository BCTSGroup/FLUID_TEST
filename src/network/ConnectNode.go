package Network

import (
	"encoding/gob"
	"errors"
	"net"
	"sync"
	"time"

	"DAG-Exp/src/account"
	"DAG-Exp/src/block"
	"DAG-Exp/src/utils"
)

type Conns struct {
	conn          net.Conn
	PtrEncoder    *gob.Encoder
	RemoteAddress string
}

var GConns []Conns
var GConnsRWlock sync.RWMutex

var GAddressConns map[string]Conns
var GAddressConnsRWLock sync.RWMutex

var GConnsIpServerPort map[string]string

func HandleConnection(conn Conns, dec *gob.Decoder) {
	//END_PEER_CONNECTION:
	selfAccount := account.TransferToAccountInDb(*account.GAccount)
	err := SendMessage(block.ConnectBasedOnNodeIpMessage(selfAccount, string(account.GAccount.Address), GPort), conn.PtrEncoder)
	if err != nil {
		utils.Log.Infof("Connecting to NodeIp  :  send selfAccount and addr Failed %s", err.Error())
	}

	P2pHandleConnection(conn, dec)
}

func Connect(peerAddr string, p2p_listen_port string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, time.Second*2)

	if err != nil {
		utils.Log.Errorf("Connect: %s : Dial Failed, error: %s", peerAddr, err.Error())
		return nil, errors.New("connect failed")
	}

	utils.Log.Errorf("Connect : Connecting  to  %s", peerAddr)

	return conn, err
}
