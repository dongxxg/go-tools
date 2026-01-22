package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/core/log"
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

func NewAsyncProducer(config *Config) (sarama.AsyncProducer, error) {
	if err := config.initSarama(); err != nil {
		return nil, errors.Wrapf(err, "initSaram")
	}
	conf := config.saram
	conf.Producer.Return.Successes = false
	conf.Producer.Return.Errors = false

	return sarama.NewAsyncProducer(strings.Split(config.Brokers, ","), config.saram)
}

type IConfigGetter interface {
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	Set(key string, val interface{})
	GetStringWithDefault(key string, v string) string
	GetIntWithDefault(key string, v int) int
	GetInt64WithDefault(key string, v int64) int64
	GetInt32WithDefault(key string, v int32) int32
	GetBoolWithDefault(key string, v bool) bool
	IsNacosSync() bool
	NacosGroup() string
}

func NewAsyncProducerWith(ctx kctx.Ctx, getter IConfigGetter, prefix string) (sarama.AsyncProducer, error) {
	conf := NewKafkaConfig(getter, prefix)
	SetLogger(&ctx)
	return NewAsyncProducer(conf)
	//return NewProducer()
	//return NewProducer(conf)
}

// SendMessage sends message to kafka topic.
// Kafka uses the key of the message to select the partition of the topic it writes to
func (p *Producer) SendMessage(topic, key, value string) error {
	var k sarama.Encoder
	if key == "" {
		k = nil
	} else {
		k = sarama.StringEncoder(key)
	}

	partition, offset, err := p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   k,
		Value: sarama.StringEncoder(value),
	})

	if err != nil {
		log.Errorf("failed to send message %s to %s: %v", value, topic, err)
		return err
	}
	log.Tracef("wrote message %s at partition: %d, offset: %d", value, partition, offset)
	return nil
}

// SendMessageWithCtx sends message to kafka topic.
// Kafka uses the key of the message to select the partition of the topic it writes to
func (p *Producer) SendMessageWithCtx(ctx kctx.Ctx, topic, key, value string) error {
	var k sarama.Encoder
	if key == "" {
		k = nil
	} else {
		k = sarama.StringEncoder(key)
	}

	partition, offset, err := p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   k,
		Value: sarama.StringEncoder(value),
	})

	if err != nil {
		//log.Errorf("failed to send message %s to %s: %v", value, topic, err)
		return err
	}
	ctx.Infof("wrote message %s at partition: %d, offset: %d", value, partition, offset)
	return nil
}

type MsgCap struct {
	Content string
	Key     string
	Topic   string
}

func (p *Producer) SendMessages(msgs []MsgCap) error {

	msg := make([]*sarama.ProducerMessage, 0, len(msgs))
	for _, v := range msgs {
		if v.Content == "" {
			continue
		}
		var k sarama.Encoder
		if v.Key == "" {
			k = nil
		} else {
			k = sarama.StringEncoder(v.Key)
		}

		msg = append(msg, &sarama.ProducerMessage{
			Topic: v.Topic,
			Key:   k,
			Value: sarama.StringEncoder(v.Content),
		})
	}

	return p.syncProducer.SendMessages(msg)
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

// OriginSendMessage sends message to kafka topic.
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
	partition, offset, err := p.syncProducer.SendMessage(m)

	if err != nil {
		return fmt.Errorf("failed to send message %s to %s: %v", msg.Value, msg.Topic, err)

	}
	log.Tracef("wrote message %s at partition: %d, offset: %d", msg.Value, partition, offset)
	return nil
}

// OriginSendMessageWithCtx sends message to kafka topic.
// Kafka uses the key of the message to select the partition of the topic it writes to
func (p *Producer) OriginSendMessageWithCtx(ctx kctx.Ctx, msg SyncSendMMessage) error {

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
	partition, offset, err := p.syncProducer.SendMessage(m)

	if err != nil {
		return errors.Wrapf(err, "send to %s failed", msg.Topic)

	}
	ctx.Debugf("wrote message partition: %d, offset: %d", partition, offset)
	log.Tracef("wrote message %s at partition: %d, offset: %d", msg.Value, partition, offset)
	return nil
}

// Cleanup should be called when producer is not used
func (p *Producer) Cleanup() error {
	return p.syncProducer.Close()
}
