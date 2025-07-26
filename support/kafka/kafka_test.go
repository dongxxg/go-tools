package kafka

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func consumeMsg(topic string, timestamp time.Time, value []byte) {
	fmt.Printf(
		"Message claimed: value = %s, timestamp = %v, topic = %s\n",
		string(value),
		timestamp,
		topic,
	)
}

func ExampleNewConsumer() {
	c, err := NewConsumer(NewConsumerConfig("192.168.118.220:9092", "test_topic", true))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	err = c.Consume(consumeMsg)
	fmt.Printf("%v\n", err)
}

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TestProducer(t *testing.T) {
	kafkaConfig := &Config{
		Version:    "3.7.0",
		Brokers:    "localhost:9092",
		SaslEnable: false,
		//UserName:   ,
		//Password:   conf.GetString("Kafka.Password"),
		//Algorithm:  conf.GetString("Kafka.Algorithm"),
		Group:    "dx_group",
		Topics:   "test_topic",
		Assignor: "range",
	}

	//producer
	p, err := NewProducer(kafkaConfig)
	if err != nil {
		return
	}
	defer p.Cleanup()

	var items []Item
	// 每个Item大约1KB，2000个Item约2M
	for i := 0; i < 2000; i++ {
		items = append(items, Item{
			ID:    i,
			Name:  strings.Repeat("test", 200), // 800字节
			Value: strings.Repeat("x", 700),    // 700字节
		})
	}

	value, _ := json.Marshal(items)

	//err = p.SendMessage("test_topic", "1111", "dx")
	// 2M消息
	err = p.SendMessage("test_topic", "1.1.1.1", string(value))
	if err != nil {
		t.Error("send msg err", err.Error())
	}
}
