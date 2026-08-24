package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"runtime"
)

func readOsStdin()  {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		fmt.Println("your input:", input.Text())
	}
}

func readFile(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Open file failed", err)
		return
	}
	input := bufio.NewScanner(file)
	for input.Scan() {
		fmt.Println("The file content line:", input.Text())
	}
}

func getCurrentDirPath() string {
	var currentPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		currentPath = path.Dir(filename)
	}
	return currentPath
}

var (
	firstName, lastName, s string
	i int
	f float32
	input = "56.12 / 5212 / Go"
	format = "%f / %d / %s"
)

func testScan(){
	fmt.Println("Please input your firstname and lastname:")
	fmt.Scanln(&firstName, &lastName) //扫描来自标准输入的文本，将空格分隔的值依次存放到后续的参数内，直到碰到换行
	fmt.Println("Hi", firstName, lastName)

	fmt.Sscanf(input, format, &f, &i, &s) // 从字符串读取
	fmt.Println("Read:", f, i, s)

	/*
		Please input your firstname and lastname:
		dong xiang
		Hi dong xiang
		Read: 56.12 5212 Go
	*/
}

func TestBufio() {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter you info:")
	input2, err := inputReader.ReadString('\n')
	if err == nil {
		fmt.Printf("Your input was: %s\n", input2)
	}
}


func main() {
	currentPath := getCurrentDirPath()
    fileName := currentPath + "\\test.txt"
    readFile(fileName)
}
