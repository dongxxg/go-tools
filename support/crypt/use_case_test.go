package crypt

import (
	"testing"
)

func TestStorageEncrypt(t *testing.T) {
	RS, _ := CommonEncrypt(
		[]byte(
			"{\"appId\":\"rk\",\"channelId\":\"123456\",\"loginName\":\"t2admin2\",\"userType\":\"host\",\"expire\":2000000000000}",
		),
		1,
		[]byte("dt47593e1156711b5a185216400f3126"),
	)
	t.Log(string(RS))
	rs2, _ := CommonDecrypt(RS, 1, []byte("dt47593e1156711b5a185216400f3126"))
	t.Log(string(rs2))
}
