package main

import "encoding/base64"

func Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	return string(data), err
}

/*
func main(){
	s := "shane"
	encode := base64.Encode(s)
	fmt.Println(encode)

	decode, err := base64.Decode(encode)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(decode)
}
*/
