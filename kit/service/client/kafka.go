package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/panjf2000/ants"
	"github.com/pkg/errors"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/conf"
)

type KafkaArg struct {
	// RecvSize 接收topic 的协程数
	RecvSize int
	// WorkSize 处理逻辑的协程数
	WorkSize        int
	Group           string // 指定消费组
	GroupFunc       func(group, nacosPort string) (string, error)
	IsGroupFuncOpen bool // 是否打开 消费组生成函数
	Arg             map[string]string
}

// NewKafkaConsumerCli 启动并运行kafka 消息者
func NewKafkaConsumerCli(ctx kctx.Ctx, conf *kafka.Config, handleFunc model.HandleFunc, arg KafkaArg) (cf func(), err error) {
	b, _ := json.Marshal(conf)
	ctx.Debugf("kafka config %s", string(b))
	kafka.SetLogger(&ctx)
	logger := sarama.Logger
	logger.Println("here wiken")
	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		return func() {

		}, errors.Wrapf(err, "new kafka consumer failed")
	}
	if arg.RecvSize == 0 {
		arg.RecvSize = 5
	}

	wg := sync.WaitGroup{}
	//workPool, err := ants.NewPool(arg.WorkSize)
	workPool, err := ants.NewPoolWithFunc(arg.WorkSize, func(i interface{}) {
		msg, ok := i.(model.ConsumerMsg)
		if !ok {
			return
		}
		handleFunc(ctx, msg)
		wg.Done()
	})
	if err != nil {
		return func() {
		}, fmt.Errorf("new work pool failed, err: %v", err)
	}

	recvInfoChan := make(chan model.ConsumerMsg, arg.RecvSize*2)

	for i := 0; i < arg.RecvSize; i++ {
		go func() {
			for info := range recvInfoChan {
				wg.Add(1)
				if err := workPool.Invoke(info); err != nil {
					ctx.Error("invoke to kafka failed, err: ", err)
				}
			}
		}()
	}
	cf_, err := consumer.Consume(ctx, func(ctx context.Context, msg model.ConsumerMsg) {
		recvInfoChan <- msg
	})
	if err != nil {
		return nil, err
	}
	return func() {
		workPool.Release()
		cf_()
	}, nil
}

type Kafka struct {
	getter     IConfigGetter
	prefix     string // 配置前缀
	handleFunc model.HandleFunc
	ctx        kctx.Ctx
	arg        KafkaArg
	configList *gset.StrSet
}

// NewKafka
/*
prefix: 配置的前缀， 如xxx.kafka 为前缀

# kafka 的配置字段如下
[xxx.kafka]
Brokers = "127.0.0.1:9092"
Oldest = false
Group = "uni-alarm-process"
Version = "2.0.0"
Assignor = "range" # 负载均衡策略  sticky  roundrobin range
SaslEnable = false
UserName = ""
Password = ""
Algorithm = "" # plain oauthbearer  sha512 sha256 gssapi
CertFile = ""
CaFile = ""
KeyFile = ""
VerifySSL = false
Realm = ""
ServiceName = ""
KeyTabPath = ""
KerberosConfigPath = ""
UseTLS
*/
func NewKafka(ctx kctx.Ctx, getter IConfigGetter, prefix string, arg KafkaArg) *Kafka {
	if arg.RecvSize == 0 {
		arg.RecvSize = 1
	}
	if arg.WorkSize == 0 {
		arg.WorkSize = 50
	}
	kafka.SetLogger(&ctx)

	gen := conf.NewConfigKey(prefix)

	return &Kafka{
		getter: getter,
		prefix: prefix,
		ctx:    ctx,
		arg:    arg,
		configList: gset.NewStrSetFrom([]string{
			gen.Key("Brokers"),
			gen.Key("Oldest"),
			gen.Key("Group"),
			gen.Key("Version"),
			gen.Key("Assignor"),
			gen.Key("SaslEnable"),
			gen.Key("UserName"),
			gen.Key("Password"),
			gen.Key("Algorithm"),
			gen.Key("CertFile"),
			gen.Key("CaFile"),
			gen.Key("KeyFile"),
			gen.Key("VerifySSL"),
			gen.Key("Realm"),
			gen.Key("ServiceName"),
			gen.Key("KeyTabPath"),
			gen.Key("KerberosConfigPath"),
		}),
	}
}
func (k *Kafka) SetConfKeys(names []string) {
	k.configList = gset.NewStrSetFrom(names)
}

func (k *Kafka) Consume(topicKey string, f model.HandleFunc) (cf func(), err error) {
	k.handleFunc = f

	if !k.getter.IsNacosSync() {
		// 不需要动态更新nacos 配置
		kafkaConf := NewConsumerKafkaConf(k.getter, k.getter.GetString(topicKey), k.prefix)

		// 开启自动生成消费组名称
		if kafkaConf.Group != "" && k.arg.GroupFunc != nil && k.arg.IsGroupFuncOpen {
			g, err := k.arg.GroupFunc(kafkaConf.Group, k.getter.GetString("Server.Http.Port"))
			if err != nil {
				return func() {}, errors.Wrapf(err, "gen consumer group")
			}
			kafkaConf.Group = g
		}

		cf, err = NewKafkaConsumerCli(k.ctx, kafkaConf, k.handleFunc, k.arg)
		return
	}

	changed := make(chan struct{})

	k.configList.Add(topicKey)

	k.configList.Iterator(func(v string) bool {
		if err_ := ListenNacos(k.getter, v, func(namespace, group, dataId, data string) {

			k.getter.Set(v, data)
			changed <- struct{}{}
		}); err_ != nil {
			k.ctx.Error("ListenNacos failed %v", err_)
		}
		return true
	})

	go func() {
		for {
			select {
			case <-k.ctx.Done():
				return
			case _, ok := <-changed:
				if !ok {
					return
				}
				kafkaConf := NewConsumerKafkaConf(k.getter, k.getter.GetString(topicKey), k.prefix)
				if cf != nil {
					cf()
				}
				cf, err = NewKafkaConsumerCli(k.ctx, kafkaConf, k.handleFunc, k.arg)
				if err != nil {
					k.ctx.Errorf("NewKafkaConsumerCli failed: %v", err)
				}
			}
		}
	}()
	changed <- struct{}{}

	return
}

func (k *Kafka) Producer() (sender ISender, cf func(), err error) {

	if !k.getter.IsNacosSync() {
		// 不需要动态更新nacos 配置
		cfg := NewProducerConf(k.getter, k.prefix)
		var p *kafka.Producer
		p, err = kafka.NewProducer(cfg)
		if err != nil {
			k.ctx.Errorf("kafka NewProducer failed: %v", err)
		}
		if sender != nil {
			sender.Set(p)
		}
		cf = func() {
			_ = p.Cleanup()
		}
		return
	}

	changed := make(chan struct{})
	k.configList.Iterator(func(v string) bool {

		if err_ := ListenNacos(k.getter, v, func(namespace, group, dataId, data string) {
			k.getter.Set(v, data)
			changed <- struct{}{}
		}); err_ != nil {
			k.ctx.Error("ListenNacos failed %v", err_)
		}
		return true
	})
	c := make(chan struct{})
	once := sync.Once{}
	go func() {
		for {
			select {
			case _, ok := <-changed:
				if !ok {
					return
				}

				cfg := NewProducerConf(k.getter, k.prefix)
				k.ctx.Infof("kafka config for producer %+v", *cfg)

				var p *kafka.Producer

				p, err = kafka.NewProducer(cfg)
				if err != nil {
					k.ctx.Errorf("kafka NewProducer failed: %v", err)
				}
				if sender != nil {
					sender.Set(p)
				} else {
					sender = &MyKafka{
						P: p,
					}
				}
				cf = func() {
					_ = p.Cleanup()
				}

				once.Do(func() {
					c <- struct{}{}
				})
			}
		}
	}()

	changed <- struct{}{}
	<-c
	return
}
func NewConsumerKafkaConf(getter IConfigGetter, topic, prefix string) *kafka.Config {
	c := NewProducerConf(getter, prefix)
	c.Topics = topic
	return c
}

func NewProducerConf(getter IConfigGetter, prefix string) *kafka.Config {
	gen := conf.NewConfigKey(prefix)

	return &kafka.Config{
		Version:            getter.GetString(gen.Key("Version")),
		Brokers:            getter.GetString(gen.Key("Brokers")),
		Group:              getter.GetString(gen.Key("Group")),
		Assignor:           getter.GetString(gen.Key("Assignor")),
		Oldest:             getter.GetBool(gen.Key("Oldest")),
		SaslEnable:         getter.GetBool(gen.Key("SaslEnable")),
		UserName:           getter.GetString(gen.Key("UserName")),
		Password:           getter.GetString(gen.Key("Password")),
		Algorithm:          getter.GetString(gen.Key("Algorithm")),
		CertFile:           getter.GetString(gen.Key("CertFile")),
		CaFile:             getter.GetString(gen.Key("CaFile")),
		KeyFile:            getter.GetString(gen.Key("KeyFile")),
		VerifySSL:          getter.GetBool(gen.Key("VerifySSL")),
		Realm:              getter.GetString(gen.Key("Realm")),
		ServiceName:        getter.GetString(gen.Key("ServiceName")),
		KeyTabPath:         getter.GetString(gen.Key("KeyTabPath")),
		KerberosConfigPath: getter.GetString(gen.Key("KerberosConfigPath")),
		Backoff:            getter.GetString(gen.Key("Backoff")),
		UseTLS:             getter.GetBool(prefix + "UseTLS"),
	}
}
