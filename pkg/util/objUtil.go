package util

import "strconv"

// 解析mongo 的_id值
// 前8位16进制的以秒计算的时间值，转化成ms
func GetObjectTime(recordId string) int64 {

	dataStr := recordId[0:8]
	result, _ := strconv.ParseInt(dataStr, 16, 64)
	return result * 1000
}
