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

func decode(src string) string {
	s, _ := hex.DecodeString(src)
	return string(s)
}

func main() {
	s := Md5("1235ggdd444")
	//856e77232f4a9398ce1edee7d5077988
	//856e77232f4a9398ce1edee7d5077988
	fmt.Println(s)

	res := decode("efdb386014092f63410cefe80dd85160")
	fmt.Println(res)
}
