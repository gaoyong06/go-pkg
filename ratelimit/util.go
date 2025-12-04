package ratelimit

import "time"

// getNowUnix 获取当前时间戳（秒）
// 提取为函数方便测试时 mock
var getNowUnix = func() int64 {
	return time.Now().Unix()
}
