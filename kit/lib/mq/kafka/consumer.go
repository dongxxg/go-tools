package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"strings"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/util"
	"unitechs.com/unios-dice/uni-base/core/log"
)

type Consumer struct {
	config *Config

	ready chan bool

	cb func(ctx context.Context, msg model.ConsumerMsg)
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

type Algorithm string

const (
	AlgorithmSha512 Algorithm = "sha512"
	AlgorithmSha256 Algorithm = "sha256"
	AlgorithmGssapi Algorithm = "gssapi"
)

// ParamKafka 需要兼容
type ParamKafka struct {
	Broker    string    `json:"kafkaBroker"`
	Topic     string    `json:"kafkaTopic"`
	User      string    `json:"kafkaUser"`
	Pass      string    `json:"kafkaPass"`
	Algorithm Algorithm `json:"algorithm"`

	//KafkaKeytab   string        `json:"kafkaKeytab"`   // base64 加密的文件
	//KafkaKrb5Conf string        `json:"kafkaKrb5conf"` // base64 加密的文件
	//Header string `json:"kafkaHeader"` // 消息头， json 字符串
	Group string `json:"kafkaGroup"`
}

func NewConsumerWith(p ParamKafka) (*Consumer, error) {
	c := NewConsumerConfig(
		p.Broker,
		p.Topic,
		false,
	)
	c.Version = "2.0.0"
	c.Group = p.Group
	if p.User != "" || p.Pass != "" {
		c.SaslEnable = true
		c.UserName = p.User
		c.Password = p.Pass
		c.Algorithm = string(p.Algorithm)
	}
	return NewConsumer(c)

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
		config.Group = util.NewUUIDStr()
		log.Warnf("no Kafka consumer group defined, create a random one: %s", config.Group)
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

// Consume will use the input function as callback to consume kafka message
func (consumer *Consumer) Consume(ctx kctx.Ctx, cb model.HandleFunc) (func() error, error) {
	client, err := sarama.NewConsumerGroup(strings.Split(consumer.config.Brokers, ","), consumer.config.Group, consumer.config.saram)
	if err != nil {
		return nil, errors.Errorf("Error creating consumer group client: %v", err)
	}
	c, cf := context.WithCancel(ctx.Context)

	consumer.cb = cb

	go func() {
		for {
			select {
			case <-c.Done():
				return
			default:
				// `Consume` should be called inside an infinite loop, when a
				// server-side rebalance happens, the consumer session will need to be
				// recreated to get the new claims
				if err1 := client.Consume(ctx, strings.Split(consumer.config.Topics, ","), consumer); err1 != nil {
					ctx.Errorf("Error from consumer: %v", err1)
				}
				// check if context was canceled, signaling that the consumer should stop
				if ctx.Err() != nil {
					return
				}
				consumer.ready = make(chan bool)
			}

		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Info("Sarama consumer up and running!...")

	return func() error {
		cf()
		return client.Close()
	}, nil
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

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		session.Context()
		log.Tracef("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		consumer.cb(session.Context(), model.ConsumerMsg{
			Timestamp:      message.Timestamp,
			BlockTimestamp: message.BlockTimestamp,
			Key:            message.Key,
			Value:          message.Value,
			Topic:          message.Topic,
			Partition:      message.Partition,
			Offset:         message.Offset,
		})
		session.MarkMessage(message, "")
	}
	return nil
}
