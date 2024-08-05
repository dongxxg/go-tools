package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
)

func CreateFile(filePath string) {
	var _, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		var file, err = os.Create(filePath)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer file.Close()
	}
	fmt.Println("===done creating file", filePath)
}

func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("存在同名文件")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetRootPath() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func ReadFile() {
	path, err := os.Getwd()
	file, err := os.Open(path + "\\pkg\\file\\test.txt")
	if err != nil {
		fmt.Printf("There occurred an err on opening the file:%v", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		fmt.Println("The input was:", line)
		if err == io.EOF {
			return
		}
	}
}

func IoutilRead() {
	path, err := os.Getwd()
	filePath := path + "\\pkg\\file\\test.txt"
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
		// panic(err.Error())
	}
	fmt.Println(buffer)
	err = ioutil.WriteFile(filePath, buffer, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteFile() {
	path, err := os.Getwd()
	filePath := path + "\\pkg\\file\\test.txt"
	outputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)
	str := "hello world\n"
	for i := 0; i < 5; {
		i++
		writer.WriteString(str)
	}
	writer.Flush()
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

func ReadFileAndParse() {
	path, err := os.Getwd()
	file, err := os.Open(path + "\\pkg\\file\\test.txt")
	if err != nil {
		fmt.Printf("There occurred an err on opening the file:%v", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	reader.ReadString()
}

func main() {
	ReadFile()
}
