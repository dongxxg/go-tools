package main

import (
	"fmt"
	"gopkg.in/antage/eventsource.v1"
	"log"
	"net/http"
	"time"
)

func main() {
	es := eventsource.New(nil, nil)
	defer es.Close()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/events", es)
	go func() {
		for {
			// 只设置发送数据，不添加事件名
			es.SendEventMessage(fmt.Sprintf("send data: %s", time.Now().Format("2006-01-02 15:04:05")), "test-event", "")
			log.Printf("客户端连接数: %d", es.ConsumersCount())
			time.Sleep(2 * time.Second)
		}
	}()

	log.Println("Open URL http://localhost:8081/ in your browser.")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
