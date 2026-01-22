package log2kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	"unitechs.com/cno/kit/service/client"
	"unitechs.com/cno/kit/service/conf"
)

var once = sync.Once{}

// NewDefaultContext new default with send to kafka
func NewDefaultContext() kctx.Ctx {
	return newContext(context.TODO(), newUUIDStr())
}

func newContext(cc context.Context, tid string) kctx.Ctx {
	c := kctx.NewContext(cc, tid)
	getter := conf.NewKitConf(viper.GetViper())
	ctx, err := NewCtxWithKafka2(c.Context, getter, tid)
	if err != nil {
		c.Fatal(err)
	}
	v := viper.GetViper()
	if v != nil {
		gtr := conf.NewKitConf(v)
		name := gtr.GetString("SD.Name")
		ctx = ctx.WithField("sname", name)
		// 获取本机ip
		ctx = ctx.WithField("sip", getIP())
	}
	return ctx
}

var ipv4 = ""

func getIP() string {
	if ipv4 != "" {
		return ipv4
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Failed to get interfaces:", err)
		return ""
	}

	// 遍历接口，打印其IPv4地址
	for _, iface := range interfaces {
		// 只考虑活动的接口
		if iface.Flags&net.FlagUp == 0 {
			continue // 接口未启用，跳过
		}

		// 遍历接口的所有地址
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Failed to get addresses:", err)
			continue
		}

		for _, addr := range addrs {
			// 只关心 IPv4 地址
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 检查是否是IPv4
			if ip.To4() != nil {
				if !ip.IsLoopback() {
					ipv4 = ip.String()
					return ipv4
				}
			}
			//if ip.To16() != nil {
			//	if !ip.IsLoopback() {
			//		return ip.String()
			//	}
			//}
		}
	}
	return ""
}

// NewCtx  设置全局的
func NewCtx(c context.Context, tid string) {
	ctx := newContext(c, tid)
	kctx.SetCtx(ctx)
}
func NewContext(ctx context.Context, tid string) kctx.Ctx {
	return newContext(ctx, tid)
}

// NewCtxWithRequest  为request 请求创建ctx
func NewCtxWithRequest(ctx context.Context, r *http.Request, headerKey string) kctx.Ctx {
	tid := r.Header.Get(headerKey)
	if tid == "" {
		tid = newUUIDStr()
	}
	return NewCtxWithField(ctx, headerKey, tid)
}

func NewCtxWithField(ctx context.Context, key, val string) kctx.Ctx {
	if val == "" {
		val = newUUIDStr()
	}
	c := newContext(ctx, val)

	c = c.WithField(key, val)
	//c.TraceKey = key
	//c.TraceVal = val
	return c
}

var m = map[string]logrus.Level{
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"trace": logrus.TraceLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}
var levelList = []logrus.Level{
	logrus.InfoLevel,
	logrus.DebugLevel,
	logrus.WarnLevel,
	logrus.ErrorLevel,
	logrus.TraceLevel,
	logrus.FatalLevel,
	logrus.PanicLevel,
}
var hook logrus.Hook

func NewCtxWithKafka2(ctx context.Context, getter client.IConfigGetter, tid string) (c kctx.Ctx, err error) {
	c = kctx.NewContext(ctx, tid)
	c.Context = ctx
	once.Do(func() {
		if !getter.GetBool("Pub.Log.Enabled") {
			c.Infof("Logging.Enabled %v", false)
			return
		}
		ls := getter.GetString("Pub.Log.Level")
		levels := []logrus.Level{}
		if strings.Contains(ls, ",") {
			for _, v := range strings.Split(ls, ",") {
				l, ok := m[v]
				if ok {
					levels = append(levels, l)
				}
			}
		} else {
			// 只能一个，则向下兼容
			l, ok := m[ls]
			if !ok {
				l = logrus.InfoLevel
			}
			for _, v := range levelList {
				if l >= v {
					levels = append(levels, v)
				}
			}
		}

		midType := strings.ToLower(getter.GetString("Pub.Log.MiddlwareType"))

		switch midType {
		case "kafka":
			var p sarama.AsyncProducer

			p, err = kafka.NewAsyncProducerWith(c, getter, "Pub.Log.Kafka")
			if err != nil {
				return
			}

			hook, err = NewKafkaHook(p, c.TraceVal, levels, &logrus.JSONFormatter{})
			if err != nil {
				err = errors.Wrapf(err, "new kafkaHook")
				return
			}

		case "mq":
			var p rocketmq.Producer
			p, err = client.NewMqProducerWithConfig(getter, "Pub.Log.MQ")
			if err != nil {
				return
			}
			if err = p.Start(); err != nil {
				err = errors.Wrapf(err, "rocketmq.Start")
				return
			}
			hook, err = NewMqHook(c, p, c.TraceVal, levels, &logrus.JSONFormatter{})
		default:
			c.Warn("invalid midType %q", midType)
			return
		}

	})

	topics := getter.GetString("Pub.Log.Topic")
	if len(topics) == 0 {
		c.Info("log2Kafka is ture, but topic is empty")
		return
	}

	c = c.WithField("topics", topics)
	if hook != nil {
		c.Logger.Logger.AddHook(hook)
	}
	return c, nil

}

func NewCtxWithKafka(tid string, getter client.IConfigGetter) (c kctx.Ctx, err error) {
	c = kctx.NewContext(context.TODO(), tid)

	once.Do(func() {
		if !getter.GetBool("Pub.Log.Enabled") {
			c.Infof("Logging.Enabled %v", false)
			return
		}
		ls := getter.GetString("Pub.Log.Level")
		levels := []logrus.Level{}
		if strings.Contains(ls, ",") {
			for _, v := range strings.Split(ls, ",") {
				l, ok := m[v]
				if ok {
					levels = append(levels, l)
				}
			}
		} else {
			// 只能一个，则向下兼容
			l, ok := m[ls]
			if !ok {
				l = logrus.InfoLevel
			}
			for _, v := range levelList {
				if l >= v {
					levels = append(levels, v)
				}
			}
		}

		midType := strings.ToLower(getter.GetString("Pub.Log.MiddlwareType"))

		switch midType {
		case "kafka":
			var p sarama.AsyncProducer

			p, err = kafka.NewAsyncProducerWith(c, getter, "Pub.Log.Kafka")
			if err != nil {
				return
			}

			hook, err = NewKafkaHook(p, c.TraceVal, levels, &logrus.JSONFormatter{})
			if err != nil {
				err = errors.Wrapf(err, "new kafkaHook")
				return
			}

		case "mq":
			var p rocketmq.Producer
			p, err = client.NewMqProducerWithConfig(getter, "Pub.Log.MQ")
			if err != nil {
				return
			}
			if err = p.Start(); err != nil {
				err = errors.Wrapf(err, "rocketmq.Start")
				return
			}
			hook, err = NewMqHook(c, p, c.TraceVal, levels, &logrus.JSONFormatter{})
		default:
			c.Warn("invalid midType %q", midType)
			return
		}

	})

	topics := getter.GetString("Pub.Log.Topic")
	if len(topics) == 0 {
		err = errors.New("log2Kafka is ture, but topic is empty")
		return
	}

	c = c.WithField("topics", topics)
	c.Logger.Logger.AddHook(hook)

	return c, nil
}

func newUUIDStr() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

type KafkaHook struct {
	// Id of the hook
	id string

	// Log levels allowed
	levels []logrus.Level

	// Log entry formatter
	formatter logrus.Formatter

	// sarama.AsyncProducer
	producer sarama.AsyncProducer
}

// NewKafkaHook Create a new KafkaHook.
func NewKafkaHook(producer sarama.AsyncProducer, id string, levels []logrus.Level, formatter logrus.Formatter) (*KafkaHook, error) {
	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to send log entry to kafka: %v\n", err)
		}
	}()

	hook := &KafkaHook{
		id,
		levels,
		formatter,
		producer,
	}

	return hook, nil
}

func (hook *KafkaHook) Id() string {
	return hook.id
}

func (hook *KafkaHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *KafkaHook) Fire(entry *logrus.Entry) error {
	// Check time for partition key
	var partitionKey sarama.ByteEncoder

	// Get field time
	t, _ := entry.Data["time"].(time.Time)

	// Convert it to bytes
	b, err := t.MarshalBinary()

	if err != nil {
		return err
	}

	partitionKey = sarama.ByteEncoder(b)

	// Check topics
	var topics []string

	if ts, ok := entry.Data["topics"]; ok {
		if tps, ok := ts.(string); !ok {
			return errors.New("Field topics must be []string")
		} else {
			topics = strings.Split(tps, ",")
		}
	} else {
		//return errors.New("Field topics not found")
		return nil
	}
	delete(entry.Data, "topics")

	// Format before writing
	b, err = hook.formatter.Format(entry)

	if err != nil {
		return err
	}

	value := sarama.ByteEncoder(b)

	for _, topic := range topics {
		hook.producer.Input() <- &sarama.ProducerMessage{
			Key:   partitionKey,
			Topic: topic,
			Value: value,
		}
	}

	return nil
}
