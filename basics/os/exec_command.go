package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func checkCommandExist() {
	path, err := exec.LookPath("ls")
	if err != nil {
		fmt.Printf("no cmd ls: %v\n", err)
	} else {
		fmt.Printf("ls in path: %v\n", path)
	}
}

func main() {
	cmd := exec.Command("cal") //linux中查看日历
	//显示到标准输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//或者输出到文件
	//f, err := os.OpenFile("out.txt", os.O_WRONLY|os.O_CREATE, os.ModePerm)
	//cmd.Stdout = f
	//cmd.Stderr = f

	// 获取输出
	//output, err := cmd.CombinedOutput()

	// 分别获得标准输出和标准错误
	//var stdout, stderr bytes.Buffer
	//cmd.Stdout = &stdout
	//cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Fail to exec command:%v\n", err)
	}
	fmt.Println("Exec command finished!")
}
