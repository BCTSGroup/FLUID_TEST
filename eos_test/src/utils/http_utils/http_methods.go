package http_utils

import (
	"bfc/src/utils"
	"errors"
	"strconv"
	"strings"
)

const (
	BASIC_FEE = 1000.0
)

var ERR_STORAGE_EXP_INVALID_DIGIT = errors.New("storage expression has invalid hex digit")
var ERR_STORAGE_EXP_INVALID_LENGTH = errors.New("storage expression has invalid length")

func CalculateContractFee(duration int64, fileLength int64, storageExp string) (float64, error) {

	if len(storageExp) > 100 {
		return 0, ERR_STORAGE_EXP_INVALID_LENGTH
	}

	storageExpSplitArr := strings.Split(storageExp, "")
	var storageExpSum int64 = 0
	for _, singleChar := range storageExpSplitArr {

		// parse int
		num, err := strconv.ParseInt(singleChar, 16, 64)
		if err != nil {
			storageExpSum = 0
			utils.Log.Errorf("Calculate Contract Fee, %s, storage express: %s", err, storageExp)
			return 0, ERR_STORAGE_EXP_INVALID_DIGIT
		}

		// calc storage multiplier
		if (num > 0 && num < 9) || (num > 10 && num < 15) {
			storageExpSum += num
		} else {
			storageExpSum = 0
			utils.Log.Errorf("storage expression: %s have invalid hex digit", storageExp)
			return 0, ERR_STORAGE_EXP_INVALID_DIGIT
		}

	}
	totalFee := duration * fileLength * BASIC_FEE * storageExpSum
	return float64(totalFee), nil
}
