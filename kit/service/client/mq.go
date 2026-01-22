package client

import "github.com/apache/rocketmq-client-go/v2/producer"

func NewMQSyncProducer() {
	producer.NewDefaultProducer()

}
