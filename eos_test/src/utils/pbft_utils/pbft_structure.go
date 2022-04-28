package pbft_utils

import "sync"

const (
	pbftPrePrepare_to_PrepareWaitingTime = 30
	pbftPrepare_to_CommitWaitingTime     = 30
	pbftCommit_to_FinalWaitingTime       = 30
)

//verify if this signature of message correspond the previous signature of message
var PBFTAuthPrePrepareMap map[string][]byte
var PBFTAuthPrePrepareMapRWLock sync.RWMutex

var PBFTAuthPrepareMap map[string][]byte
var PBFTAuthPrepareMapRWLock sync.RWMutex

var PBFTAuthCommitMap map[string][]byte
var PBFTAuthCommitMapRWLock sync.RWMutex

//count the consensus request is more than two third of all block producing node or not
var PBFTVoteAuthPrepareMap map[string]int
var PBFTVoteAuthPrepareMapLock sync.Mutex

var PBFTVoteAuthCommitMap map[string]int
var PBFTVoteAuthCommitMapLock sync.Mutex
