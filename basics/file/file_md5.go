package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	path, _ := os.Getwd()
	filePath := path + "\\pkg\\file\\test.txt"
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("failed to open file", err)
	}
	defer file.Close()

	hash := md5.New() // 创建 MD5 计算器
	// 将文件内容拷贝到计算器中
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println("failed to copy to md5", err)
	}
	// 将文件内容拷贝到计算器中
	md5Hash := hash.Sum(nil)
	// 将二进制的 MD5 值转换为字符串
	md5String := hex.EncodeToString(md5Hash)
	fmt.Println("file md5 value:", md5String)
}
