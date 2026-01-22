package client

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	"unitechs.com/cno/kit/util"
)

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

type ISender interface {
	Send(topic, key, content string) error
	SendWithCtx(ctx kctx.Ctx, topic, key, content string) error
	BatchSend(topic, key, content string) error
	Set(v interface{}) // 设置内核
}

// MiddlewareType 中间件类型
type MiddlewareType string

const (
	MiddlewareTypeKafka    MiddlewareType = "kafka"
	MiddlewareTypeRocketmq MiddlewareType = "rocketmq" // rocketmq或ctg-mq
)

// NewProducerCli 初始化生产者， 基于创研底座
// KafkaPrefix 配置文件前缀
// ctg-mq 配置文件前缀
func NewProducerCli(
	ctx kctx.Ctx,
	sendType MiddlewareType,
	getter IConfigGetter,
	kafkaPrefix, ctgMqPrefix string) (cf func(), sender ISender, err error) {

	switch sendType {
	case MiddlewareTypeKafka:
		return NewKafkaProducer(ctx, getter, kafkaPrefix)
	case MiddlewareTypeRocketmq:
		return NewRocketMqProducerWithConfig(ctx, getter, ctgMqPrefix)
	default:
		return nil, nil, fmt.Errorf("unsupported sendType %s", sendType)
	}
}

// NewProducerCliWithNacos new 通用的生产者client, 包括kafka/ctg-mq
// 实现nacos 热更新
func NewProducerCliWithNacos(
	ctx kctx.Ctx,
	sendType MiddlewareType,
	getter IConfigGetter,
	kafkaPrefix, ctgMqPrefix string) (cf func(), sender ISender, err error) {

	switch sendType {
	case MiddlewareTypeKafka:
		cli := NewKafka(ctx, getter, kafkaPrefix, KafkaArg{})
		sender, cf, err = cli.Producer()
		return
	case MiddlewareTypeRocketmq:
		cli := NewCtgMq(ctx, getter, ctgMqPrefix, CtgMqCliPara{})
		sender, cf, err = cli.Producer()
		return
	default:
		return nil, nil, fmt.Errorf("unsupported sendType %s", sendType)
	}
}

// NewKafkaProducer
/*
Version = "2.0.0"
Brokers = "localhost:9092" # 多个使用逗号分隔
#Topics = "unios-alarm-notify"
Group = "unios-alarm-recv-group" # 默认消费组
Assignor = "range" # 负载均衡策略  sticky  roundrobin range
Oldest = false
SaslEnable = false
UserName = ""
Password = ""
Algorithm = "" # plain oauthbearer  sha512 sha256 gssapi
# Kerberos时  algrithm 为 gssapi，
# 配置
# Realm, ServiceName, UserName
# KeyTabPath, KerberosConfigPath
CertFile = ""
CaFile = ""
KeyFile = ""
VerifySSL = false
Realm = ""
ServiceName = ""
KeyTabPath = ""
KerberosConfigPath = ""
*/
func NewKafkaProducer(ctx kctx.Ctx, getter IConfigGetter, prefix string) (cf func(), sender ISender, err error) {
	kafka.SetLogger(&ctx)
	// 前缀最后不用带点
	prefix = prefix + "."
	var p *kafka.Producer
	conf := kafka.NewKafkaConfig(getter, prefix)

	ctx.Infof("kafka config for producer %+v", *conf)
	p, err = kafka.NewProducer(conf)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "new kafka producer failed")
	}
	mk := &MyKafka{
		P:   p,
		Ctx: ctx,
	}
	mk.Run()
	return func() {
		_ = p.Cleanup()
	}, mk, err
}

func NewAsyncProducer(ctx kctx.Ctx, getter IConfigGetter, prefix string, middlewareType MiddlewareType) {
	kafka.SetLogger(&ctx)
	conf := kafka.NewKafkaConfig(getter, prefix)
	kafka.NewAsyncProducer(conf)
}

// NewRocketMqProducerWithConfig
/*
Enalbed = true
Brokers = "127.0.0.1:9876"
Group = "unios-hedge-group" # 注意，千万不能为空
Username = ""
Password = ""
# 实例名称， 非必填
InstanceName = ""
*/
func NewRocketMqProducerWithConfig(ctx kctx.Ctx, getter IConfigGetter, prefix string) (cf func(), sender ISender, err error) {
	broker := getter.GetString(prefix + ".Brokers")
	p, err := NewMqProducerWithConfig(getter, prefix)
	if err != nil {
		return nil, nil, errors.Wrap(err, "NewMqProducerWithConfig failed")
	}
	if err = p.Start(); err != nil {
		return nil, nil, errors.Wrapf(err, "start ctg-mq producer failed, broker %s", broker)
	}

	mq := &MyRocketMq{
		P:   p,
		Ctx: ctx,
	}
	mq.Run()
	return func() {
		_ = p.Shutdown()
	}, mq, nil
}
func NewMqProducerWithConfig(getter IConfigGetter, prefix string) (rocketmq.Producer, error) {
	retry := getter.GetInt(prefix + "retry")
	if retry == 0 {
		retry = 2
	}
	instanceName := getter.GetString(prefix + "instanceName")
	if instanceName == "" {
		instanceName = util.NewUUIDStr()
	}

	ops := make([]producer.Option, 0)

	broker := getter.GetString(prefix + ".Brokers")
	ops = append(ops, producer.WithNsResolver(primitive.NewPassthroughResolver(strings.Split(broker, ","))))
	ops = append(ops, producer.WithRetry(retry))
	ops = append(ops, producer.WithInstanceName(instanceName))
	ops = append(ops, producer.WithQueueSelector(producer.NewHashQueueSelector()))
	user := getter.GetString(prefix + ".Username")
	pwd := getter.GetString(prefix + ".Password")
	if user != "" && pwd != "" {
		ops = append(ops, producer.WithCredentials(primitive.Credentials{
			AccessKey: user,
			SecretKey: pwd,
		}))
	}
	p, err := rocketmq.NewProducer(ops...)
	if err != nil {
		return nil, errors.Wrapf(err, "new ctg-mq producer failed")
	}
	return p, nil
}

type MyRocketMq struct {
	P                  rocketmq.Producer
	once               sync.Once
	SendIntervalSecond int
	msgChan            chan kafka.MsgCap
	Ctx                kctx.Ctx
}

func (m *MyRocketMq) Run() {
	m.msgChan = make(chan kafka.MsgCap)
	go func() {
		if m.Ctx.TraceVal == "" {
			m.Ctx = kctx.NewDefaultContext()
		}

		if m.SendIntervalSecond == 0 {
			m.SendIntervalSecond = 1
		}
		tick := time.NewTicker(time.Second * time.Duration(m.SendIntervalSecond))
		msg := make([]kafka.MsgCap, 0, 100)
		for {
			select {
			case info, ok := <-m.msgChan:
				if !ok {
					return
				}
				msg = append(msg, info)
			case <-tick.C:
				if len(msg) > 0 {
					if err := m.SendMessages(msg); err != nil {
						m.Ctx.Errorf("mq send failed: %v", err)
						msg = make([]kafka.MsgCap, 0, 100)
					}
					msg = make([]kafka.MsgCap, 0, 100)
				}
			}
		}
	}()
}

func (m *MyRocketMq) BatchSend(topic, key, content string) error {

	m.msgChan <- kafka.MsgCap{
		Content: content,
		Key:     key,
		Topic:   topic,
	}
	return nil
}
func (m *MyRocketMq) SendMessages(msgs []kafka.MsgCap) error {
	ms := make([]*primitive.Message, 0, len(msgs))
	var topic string
	for _, v := range msgs {
		topic = v.Topic
		temp := primitive.NewMessage(v.Topic, []byte(v.Content))
		temp.WithProperty(primitive.PropertyShardingKey, v.Key)
		ms = append(ms, temp)
	}
	_, err := m.P.SendSync(context.TODO(), ms...)
	if err != nil {
		return errors.Wrapf(err, "send to ctg-mq %s failed", topic)
	}
	return nil
}

func (m *MyRocketMq) Set(v interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *MyRocketMq) Send(topic, key, content string) error {
	msg := primitive.NewMessage(topic, []byte(content))
	msg.WithProperty(primitive.PropertyShardingKey, key)
	_, err := m.P.SendSync(context.TODO(), msg)
	if err != nil {
		return errors.Wrapf(err, "send to ctg-mq failed")
	}

	return nil
}
func (m *MyRocketMq) SendWithCtx(ctx kctx.Ctx, topic, key, content string) error {
	msg := primitive.NewMessage(topic, []byte(content))
	msg.WithProperty(primitive.PropertyShardingKey, key)
	r, err := m.P.SendSync(context.TODO(), msg)
	if err != nil {
		return errors.Wrapf(err, "send to ctg-mq failed")
	}
	if r != nil {
		ctx.Infof("send info %s", r.String())
	}

	return nil
}

type MyKafka struct {
	P                  *kafka.Producer
	l                  sync.RWMutex
	SendIntervalSecond int
	msgChan            chan kafka.MsgCap
	once               sync.Once
	Ctx                kctx.Ctx
}

func (m *MyKafka) Run() {
	m.msgChan = make(chan kafka.MsgCap)
	go func() {
		if m.SendIntervalSecond == 0 {
			m.SendIntervalSecond = 1
		}

		tick := time.NewTicker(time.Second * time.Duration(m.SendIntervalSecond))
		msg := make([]kafka.MsgCap, 0, 100)
		for {
			select {
			case info, ok := <-m.msgChan:
				if !ok {
					return
				}
				msg = append(msg, info)
			case <-tick.C:
				if len(msg) > 0 {
					if err := m.P.SendMessages(msg); err != nil {
						m.Ctx.Errorf("batch send to kafka %s failed: %v", msg[0].Topic, err)
						msg = make([]kafka.MsgCap, 0, 100)
						continue
					}
					m.Ctx.Infof("send to kafka %s success len %d", msg[0].Topic, len(msg))
					msg = make([]kafka.MsgCap, 0, 100)
				}
			}
		}
	}()
}
func (m *MyKafka) BatchSend(topic, key string, content string) error {

	m.msgChan <- kafka.MsgCap{
		Content: content,
		Key:     key,
		Topic:   topic,
	}
	return nil
}

func (m *MyKafka) Set(v interface{}) {
	m.l.Lock()

	if p, ok := v.(*kafka.Producer); ok {
		_ = m.P.Cleanup()
		m.P = p
	}
	m.l.Unlock()
}

func (m *MyKafka) Send(topic, key, content string) error {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.P.SendMessage(topic, key, content)
}
func (m *MyKafka) SendWithCtx(ctx kctx.Ctx, topic, key, content string) error {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.P.SendMessageWithCtx(ctx, topic, key, content)
}
