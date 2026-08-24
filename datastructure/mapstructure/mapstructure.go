package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type Blog struct {
	BlogId  string `mapstructure:"blogId"`
	Title   string `mapstructrue:"title"`
	Content string `mapstructure:"content"`
	Uid     string `mapstructure:"uid"`
	State   string `mapstructure:"state"`
}

type Event struct {
	Type     string              `json:"type"`
	Database string              `json:"database"`
	Table    string              `json:"table"`
	Data     []map[string]string `json:"data"`
}

func main() {
	e := Event{}
	msg := []byte(
		`{"type": "UPDATE", "database": "blog", "table": "blog",
         "data": [{"blogId": "100001", "title": "title", 
         "content": "this is a blog", "uid": "1000012",
         "state": "1"}]}`)
	err := json.Unmarshal(msg, &e)
	if err != nil {
		fmt.Println(err)
	}
	if e.Table == "blog" {
		var blogs []Blog
		if err := mapstructure.Decode(e.Data, &blogs); err != nil {
			fmt.Println(err)
		}
		fmt.Println(blogs)
	}
}
