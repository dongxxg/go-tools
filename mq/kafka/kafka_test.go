package kafka

import (
	"fmt"
	"time"

	"github.com/dongxxg/go-tools/lib/stringx"
)

func consumeMsg(topic string, timestamp time.Time, value []byte) {
	fmt.Printf("Message claimed: value = %s, timestamp = %v, topic = %s\n", string(value), timestamp, topic)
}

// func TestMain(m *testing.M) {
// 	goleak.VerifyTestMain(m)
// }

func ExampleNewConsumer() {
	c, err := NewConsumer(NewConsumerConfig("192.168.118.220:9092", "test_topic", true))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	err = c.Consume(consumeMsg)
	fmt.Printf("%v\n", err)
}

func ExampleNewProducer() {
	p, err := NewProducer(NewProducerConfig("192.168.118.220:9092"))
	if err != nil {
		return
	}
	defer p.Cleanup()

	for i := 0; i < 10; i++ {
		p.SendMessage("test_topic", "1.1.1.1", stringx.Randn(32))
	}
}