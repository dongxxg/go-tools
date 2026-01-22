package log2kafka

import (
	"context"
	"errors"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"time"
	"unitechs.com/cno/kit/kctx"
)

type MqHook struct {
	ctx kctx.Ctx
	// Id of the hook
	id string

	// Log levels allowed
	levels []logrus.Level

	// Log entry formatter
	formatter logrus.Formatter

	// sarama.AsyncProducer
	producer rocketmq.Producer
}

// 设置自定义回调
var callback = func(ctx kctx.Ctx) func(ctx context.Context, sendResult *primitive.SendResult, err error) {
	return func(c context.Context, sendResult *primitive.SendResult, err error) {
		if err != nil {
			ctx.Errorf("Failed to send log entry to kafka: %v\n", err)
		}
	}
}

// NewMqHook Create a new MqHook.
func NewMqHook(ctx kctx.Ctx, producer rocketmq.Producer, id string, levels []logrus.Level, formatter logrus.Formatter) (*MqHook, error) {
	hook := &MqHook{
		ctx:       ctx,
		id:        id,
		levels:    levels,
		formatter: formatter,
		producer:  producer,
	}

	return hook, nil
}

func (hook *MqHook) Id() string {
	return hook.id
}

func (hook *MqHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *MqHook) Fire(entry *logrus.Entry) error {
	t, _ := entry.Data["time"].(time.Time)

	// Convert it to bytes
	b, err := t.MarshalBinary()

	if err != nil {
		return err
	}

	// Check topics
	var topics []string

	if ts, ok := entry.Data["topics"]; ok {
		if topics, ok = ts.([]string); !ok {
			return errors.New("Field topics must be []string")
		}
	} else {
		return errors.New("Field topics not found")
	}

	// Format before writing
	b, err = hook.formatter.Format(entry)

	if err != nil {
		return err
	}

	for _, topic := range topics {
		msg := &primitive.Message{
			Topic: topic,
			Body:  b,
		}
		if err = hook.producer.SendAsync(context.TODO(), callback(hook.ctx), msg); err != nil {
			hook.ctx.Errorf("mq async send failed: %v", err)
		}
	}

	return nil
}
