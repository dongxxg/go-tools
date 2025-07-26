package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Producer struct {
	config       *Config
	syncProducer sarama.SyncProducer
}

// NewProducerConfig creates a new producer configuration
func NewProducerConfig(brokers string) *Config {
	return &Config{
		Version: DEFAULT_VERSION,
		Brokers: brokers,
	}
}

// NewProducer returns a new Kafka sync producer
func NewProducer(config *Config) (*Producer, error) {
	if config.Brokers == "" {
		return nil, errors.Errorf("no Kafka bootstrap brokers defined")
	}

	err := config.initSarama()
	if err != nil {
		return nil, err
	}

	syncProducer, err := sarama.NewSyncProducer(strings.Split(config.Brokers, ","), config.saram)
	if err != nil {
		return nil, errors.Errorf("failed to create producer: %v", err)
	}

	return &Producer{
		config:       config,
		syncProducer: syncProducer,
	}, nil
}

// SendMessage sends message to kafka topic.
// Kafka uses the key of the message to select the partition of the topic it writes to
func (p *Producer) SendMessage(topic, key, value string) error {
	_, _, err := p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	})
	if err != nil {
		//log.Tracef("failed to send message %s to %s: %v", value, topic, err)
		return err
	}
	//log.Tracef("wrote message %s at partition: %d, offset: %d", value, partition, offset)
	return nil
}

type SyncSendMMessage struct {
	Topic     string // The Kafka topic for this message.
	Key       string
	Value     string
	Headers   []Header
	Metadata  interface{}
	Timestamp time.Time
	retries   int
}
type Header struct {
	Key []byte
	Val []byte
}

// SendMessage sends message to kafka topic.
// Kafka uses the key of the message to select the partition of the topic it writes to
func (p *Producer) OriginSendMessage(msg SyncSendMMessage) error {

	m := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(msg.Key),
		Value: sarama.StringEncoder(msg.Value),
	}
	m.Headers = make([]sarama.RecordHeader, 0)
	for _, v := range msg.Headers {
		m.Headers = append(m.Headers, sarama.RecordHeader{
			Key:   v.Key,
			Value: v.Val,
		})
	}
	_, _, err := p.syncProducer.SendMessage(m)

	if err != nil {
		return fmt.Errorf("failed to send message %s to %s: %v", msg.Value, msg.Topic, err)

	}
	//log.Tracef("wrote message %s at partition: %d, offset: %d", msg.Value, partition, offset)
	return nil
}

// Cleanup should be called when producer is not used
func (p *Producer) Cleanup() error {
	return p.syncProducer.Close()
}
