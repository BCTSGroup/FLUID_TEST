package utils

import (
	"time"
)

var counter uint32 = 0
var beginTime int64
var unixTimeCache int64

//var unixTimeCacheRWLock sync.RWMutex

func init() {
	beginTime = beginEpochTime()
	go timeUpdater()
}

func timeUpdater() {
	ticker := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-ticker.C:
			unixTimeCache = time.Now().Unix()
		}
	}
}

func beginEpochTime() int64 {
	return time.Date(1987, time.November, 17, 0, 0, 0, 0, time.UTC).Unix()
}

// FIXME too many syscall
//GetEpochTime get epoch time
func GetEpochTime(times int64) int64 {

	if times == 0 {
		return unixTimeCache
	}

	return times - beginTime
}

func GetUTCTime() time.Time {
	return time.Now().UTC()
}

func GetUTCZeroClockTime() int64 {
	t := time.Now()
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return tm1.Unix()
}
