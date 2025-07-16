package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func Md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5Object(obj interface{}) string {
	d, _ := json.Marshal(obj)
	return Md5String(string(d))
}

// Md5ForPlainText 用于加密服务端直接获取到的明文密码
func Md5ForPlainText(text string) string {
	return Md5String(strings.ToLower(Md5String(text)))
}

// Md5ForCipherText 用于加密前端加密过一次的密码（无论前端是否做一次转小写这里再做一次）
func Md5ForCipherText(text string) string {
	return Md5String(strings.ToLower(text))
}
