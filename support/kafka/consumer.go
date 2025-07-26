package kafka

import (
	"context"
	"github.com/google/uuid"
	"sync"
	"time"

	"strings"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Consumer struct {
	config *Config
	client sarama.ConsumerGroup
	ready  chan bool
	cb     func(topic string, timestamp time.Time, value []byte)
}

// NewConsumerConfig creates a simplest Kafka consumer configuration
// If need to control precisely, please create Config instance directly.
func NewConsumerConfig(brokers, topics string, oldest bool) *Config {
	return &Config{
		Version:  DEFAULT_VERSION,
		Brokers:  brokers,
		Topics:   topics,
		Oldest:   oldest,
		Assignor: "range",
	}
}

// NewConsumer returns a new Kafka consumer group
func NewConsumer(config *Config) (*Consumer, error) {
	if config.Brokers == "" {
		return nil, errors.Errorf("no Kafka bootstrap brokers defined")
	}

	if config.Topics == "" {
		return nil, errors.Errorf("no topics given to be consumed")
	}

	if config.Group == "" {
		// 1个partition只能被同group的一个consumer消费，随机创建group，多个consumer可以消费同一个topic
		config.Group = uuid.New().String()
		//log.Warnf("no Kafka consumer group defined, create a random one: %s", config.Group)
	}

	err := config.initSarama()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		config: config,
		ready:  make(chan bool),
	}, nil
}

// Consumer使用输入函数作为回调函数来消费kafka消息
func (consumer *Consumer) Consume(cb func(topic string, timestamp time.Time, value []byte)) error {
	client, err := sarama.NewConsumerGroup(
		strings.Split(consumer.config.Brokers, ","),
		consumer.config.Group,
		consumer.config.saram,
	)
	consumer.client = client
	if err != nil {
		return errors.Errorf("Error creating consumer group client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err1 := client.Consume(ctx, strings.Split(consumer.config.Topics, ","), consumer); err1 != nil {
				//log.Errorf("Error from consumer: %v", err1)
				return
			}
			// check if context was canceled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // 等待消费者设置完毕
	//log.Info("Sarama consumer up and running!...")

	consumer.cb = cb

	<-ctx.Done()
	cancel()
	wg.Wait()
	/*if err = client.Close(); err != nil {
		return errors.Errorf("Error closing client: %v", err)
	}*/
	return nil
}

// 启动kafka消费者并且不进行阻塞 方便用户自己关闭kafka
func (consumer *Consumer) AsyncConsume(
	cb func(topic string, timestamp time.Time, value []byte),
) (func() error, error) {
	client, err := sarama.NewConsumerGroup(
		strings.Split(consumer.config.Brokers, ","),
		consumer.config.Group,
		consumer.config.saram,
	)
	if err != nil {
		return nil, errors.Errorf("Error creating consumer group client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err1 := client.Consume(ctx, strings.Split(consumer.config.Topics, ","), consumer); err1 != nil {
				//log.Errorf("Error from consumer: %v", err1)
			}
			// check if context was canceled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	//log.Info("Sarama consumer up and running!...")
	consumer.cb = cb
	closeFn := func() error {
		cancel()
		return client.Close()
	}
	return closeFn, nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) Close() error {
	consumer.client.Close()
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		//log.Tracef(
		//	"Message claimed: value = %s, timestamp = %v, topic = %s",
		//	string(message.Value),
		//	message.Timestamp,
		//	message.Topic,
		//)
		consumer.cb(message.Topic, message.Timestamp, message.Value)
		session.MarkMessage(message, "")
	}

	return nil
}
