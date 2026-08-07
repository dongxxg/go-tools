package util

import (
	"fmt"
	"testing"
	"time"
)

func TestMoreGetObjectId(t *testing.T) {
	// 测试 一次性10个 不重复
	for i := 0; i < 100; i++ {
		objectId := GenObjectId()
		fmt.Println(objectId)
		fmt.Println(GetTimeWithObjectId(objectId))
	}
}

func TestGetObjectIdTime(t *testing.T) {
	nowTime := time.Now().UnixMilli()
	objectId := GenObjectIdWithTime(nowTime)
	fmt.Println(GetTimeWithObjectId(objectId))
}

func TestGetObjectIdTime2(t *testing.T) {
	//根据毫秒时间戳 生成objectid
	nowTime := time.Now().UnixMilli()
	objectId := GenObjectIdWithTime(nowTime)
	fmt.Println(objectId)
}
