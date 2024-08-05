package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func Md5(str string) string {
	s := md5.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

func main() {
	s := Md5("123456")
	fmt.Println(s)
}
