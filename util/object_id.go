package util

import (
	"crypto/md5"
	"encoding/hex"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

/*
* 根据当前时间 生成的objectid
 */
func GenObjectId() string {
	return bson.NewObjectId().Hex()
}

/*
*根据objectid获取对应的时间戳
 */
func GetTimeWithObjectId(objectId string) int64 {
	return bson.ObjectIdHex(objectId).Time().UnixMilli()
}

/*
* 根据时间戳 生成对应的objectid
 */
func GenObjectIdWithTime(timestamp int64) string {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	return bson.NewObjectIdWithTime(t).Hex()
}
func GenUUID() string {
	return uuid.NewV4().String()
}

func GenUUIDWithoutHyphen() string {
	return strings.Replace(GenUUID(), "-", "", -1)
}

func GenRandomHashId() string {
	str := GenUUIDWithoutHyphen()
	// 创建一个 MD5 的哈希对象
	hash := md5.New()

	// 将字符串转换为字节数组并计算哈希值
	hash.Write([]byte(str))
	hashValue := hash.Sum(nil)

	// 将哈希值转换为十六进制字符串
	md5String := hex.EncodeToString(hashValue)
	return md5String
}
