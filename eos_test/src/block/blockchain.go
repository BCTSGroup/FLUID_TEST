package block

import (
	"sync"
)

//维护的链的主要内容和它的读写锁，记得加锁。同时间允许多个读，只能一个写， 读写加锁方式不同
var GChainMem ChainMemInNode
var GChainMemRWlock sync.RWMutex

//在内存中维护的当前最新块的信息，缓存作用，避免多次数据库io影响效率,注意多个协程使用加锁
type ChainMemInNode struct {
	Height         int64
	Hash           string
	PrevBlockHash  string
	Timestamp      int64
	NodeType       int
	ProduceAccount []string
}
