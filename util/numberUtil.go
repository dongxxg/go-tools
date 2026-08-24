package util

import "math/rand"

// FindMissingNums
//	@Description:
// 从一组递增的数字中，找到空缺的数字.数字从startNum开始找
// 空缺的数字是指递增数字中没有出现的数字
// 例如：[1,2,3,5,6,8]中，startNum 为0，count为5，那么0,4,7,9,10就是空缺的数字
//
//	@param nums 是一组递增的数字
//	@param startNum, 空缺的数字的最小值
//	@param count, 需要找到空缺数字的个数
//	@return []int64 空缺的数字

func FindMissingNums(nums []int64, startNum int64, count int64) []int64 {
	if count <= 0 {
		return []int64{}
	}
	existNum := make(map[int64]bool, len(nums))
	for _, num := range nums {
		existNum[num] = true
	}
	var missingNums []int64
	for i := startNum; ; i++ {
		if !existNum[i] {
			missingNums = append(missingNums, i)
			if count--; count == 0 {
				break
			}
		}
	}
	return missingNums
}

// GenNums
//
//		@Description: 从指定数字start开始，生成指定个数count的递增的连续的数字
//		@param startd
//	    @param count
//		@return []int64
func GenNums(start, count int64) []int64 {
	if count <= 0 {
		return []int64{}
	}
	var nums []int64
	for i := start; i <= start+count-1; i++ {
		nums = append(nums, i)
	}
	return nums
}

var charList = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

// GetNumString
//
//	@Description: 随机生成指定位数的数字字符串;允许以0开始；目前用在获取邀请码上
//	@param digit: 位数;
//	@return string:生成的字符串
func GetNumStr(digit int) string {
	if digit <= 0 {
		return ""
	}
	res := make([]byte, digit)
	for i := 0; i < digit; i++ {
		res[i] = charList[rand.Intn(len(charList))]
	}
	return string(res)
}
