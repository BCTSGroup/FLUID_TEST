package dag

import (
	"math/rand"
	"time"
)

func ProcessRequestData() bool{
	// 验证访问请求

	// 提供数据处理结果
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(100)
	if i <= 3 {
		return false
	}
	return true
}