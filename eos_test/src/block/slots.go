package block

import (
	"bfc/src/utils"
)

//GetSlotNumber get slot number
func GetSlotNumber(epochTime int64) int64 {
	if epochTime == 0 {
		epochTime = utils.GetEpochTime(0)
	}
	//math.floor向下取整
	//获取当前应该出第几个块，通过时间和出块间隙决定
	return epochTime / SLOT_TIME_INTERVAL
}
