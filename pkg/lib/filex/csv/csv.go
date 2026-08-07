package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocarina/gocsv"
	"log"
	"os"
)

type Person struct {
	Name    string `csv:"Name"`
	Age     int    `csv:"Age"`
	Country string `csv:"Country"`
}

func createCSVFile() {
	filePath := "data.csv"

	// 判断文件是否存在
	_, err := os.Stat(filePath)
	exists := !os.IsNotExist(err)

	var csvFile *os.File
	var err2 error

	if !exists {
		// 创建文件
		csvFile, err2 = os.Create(filePath)
		if err2 != nil {
			log.Fatalf("创建文件失败: %v", err2)
		}
		defer csvFile.Close()

		writer := csv.NewWriter(csvFile)
		writer.Comma = '|' // 设置分隔符为 |
		defer writer.Flush()

		// 写入数据
		data := [][]string{
			{"Name", "Age", "Country"},
			{"John Doe", "30", "USA"},
			{"Jane Smith", "25", "UK"},
		}

		for _, record := range data {
			if err := writer.Write(record); err != nil {
				log.Fatalf("写入CSV失败: %v", err)
			}
		}
	} else {
		// 文件存在，以只读方式打开（或根据需要使用写入模式）
		csvFile, err2 = os.Open(filePath)
		if err2 != nil {
			log.Fatalf("打开文件失败: %v", err2)
		}
		defer csvFile.Close()

		// 可以在这里添加读取CSV内容的逻辑
	}
}

func readCSVWithGocsv() {
	csvFile, err := os.OpenFile("data.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	var people []*Person
	if err = gocsv.UnmarshalFile(csvFile, &people); err != nil {
		panic(err)
	}

	for _, p := range people {
		// 处理每个Person对象
		fmt.Printf("%#v\n", p)
	}
}

func main() {
	createCSVFile()
	//readCSVWithGocsv()
}
