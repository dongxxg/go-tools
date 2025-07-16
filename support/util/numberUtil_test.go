package util

import (
	"fmt"
	"testing"
)

func TestFindMissingNums(t *testing.T) {
	nums := []int64{-10, -1, 2, 3, 5, 6, 8}
	res := FindMissingNums(nums, 1, 5)
	t.Log(res)

	nums2 := []int64{}
	res2 := FindMissingNums(nums2, 1, 5)
	t.Log(res2)

	res3 := FindMissingNums(nil, 1, 5)
	t.Log(res3)

	res4 := FindMissingNums(nil, -1, 1)
	t.Log(res4)

	res5 := FindMissingNums(nil, -1, -1)
	t.Log(res5)
}

func TestGenNums(t *testing.T) {
	res := GenNums(1, 10)
	t.Log(res)
}

func TestGetNumStr(t *testing.T) {
	for i := 0; i < 100; i++ {
		res := GetNumStr(6)
		fmt.Println(res)
	}
}
