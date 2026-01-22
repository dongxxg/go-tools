package client

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/panjf2000/ants"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/cno/kit/util"
)

type CtgMqCliConf struct {
	Brokers      string
	Topic        string
	UserName     string
	Password     string
	Group        string
	InstanceName string
	MessageModel string // 消息模式  BroadCasting/Clustering
}
type CtgMqCliPara struct {
	WorkSize     int    // 工作线程数
	MaxSpan      int    // span 原框架默认1000， 现场有太小的情况，可定义为2000
	MessageModel string // 消息模式  BroadCasting/Clustering
}

func NewCtgMqCli(ctx kctx.Ctx, conf CtgMqCliConf, handleFunc model.HandleFunc, arg CtgMqCliPara) (cf func(), err error) {
	b, _ := json.Marshal(conf)
	ctx.Debugf("ctg-mq config %s", string(b))

	options := make([]consumer.Option, 0)
	addr, err := primitive.NewNamesrvAddr(strings.Split(conf.Brokers, ",")...)
	if err != nil {
		return
	}
	options = append(options, consumer.WithNameServer(addr))

	options = append(options, consumer.WithGroupName(conf.Group))

	if conf.UserName != "" && conf.Password != "" {
		options = append(options, consumer.WithCredentials(primitive.Credentials{
			AccessKey: conf.UserName,
			SecretKey: conf.Password,
		}))
	}
	if arg.MaxSpan == 0 {
		arg.MaxSpan = 2000
	}
	if conf.InstanceName == "" {
		conf.InstanceName = util.NewUUIDStr()
	}
	options = append(options, consumer.WithConsumeConcurrentlyMaxSpan(arg.MaxSpan))

	options = append(options, consumer.WithInstance(conf.InstanceName))
	if conf.MessageModel == consumer.BroadCasting.String() {
		options = append(options, consumer.WithConsumerModel(consumer.BroadCasting))
	}
	com, err := rocketmq.NewPushConsumer(options...)
	if err != nil {
		ctx.Errorf("New rocketMQ consumer failed, err=%v \nwait to try again... ...", err)
		time.Sleep(time.Second * 10)
		return NewCtgMqCli(ctx, conf, handleFunc, arg)
	}
	infoChan := make(chan model.ConsumerMsg)
	workPool, err := ants.NewPool(arg.WorkSize)

	go func() {
		for v := range infoChan {
			select {
			case <-ctx.Done():
				return
			default:
				func(c kctx.Ctx, m model.ConsumerMsg) {
					if err := workPool.Submit(func() {
						handleFunc(c.Context, m)
					}); err != nil {
						ctx.Errorf("submit to work pool failed, err: %v", err)
					}
				}(ctx, v)
			}

		}
	}()

	if err = com.Subscribe(conf.Topic, consumer.MessageSelector{}, func(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, i := range ext {
			infoChan <- model.ConsumerMsg{
				Headers:   nil,
				Timestamp: time.Now(),
				Key:       []byte(i.GetKeys()),
				Value:     i.Body,
				Topic:     i.Topic,
				Partition: int32(i.Queue.QueueId),
				Offset:    i.QueueOffset,
				Arg:       nil,
			}
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		ctx.Errorf("subscribe rocketMQ topic failed, topic=%s, err=%v \nwait to try again... ...", conf.Topic, err)
		time.Sleep(time.Second * 10)
		return NewCtgMqCli(ctx, conf, handleFunc, arg)
	}
	if err = com.Start(); err != nil {
		ctx.Errorf("start rocketMQ failed, err=%v \nwait to try again... ...", err)
		time.Sleep(time.Second * 10)
		return NewCtgMqCli(ctx, conf, handleFunc, arg)

	}

	return func() {
		if err := com.Shutdown(); err != nil {
			ctx.Errorf("failed to shutdown ctgmq, err: %v", err)
		}
	}, nil
}

type CtgMq struct {
	getter     IConfigGetter
	prefix     string // 配置前缀
	handleFunc model.HandleFunc
	ctx        kctx.Ctx
	arg        CtgMqCliPara
	configList *gset.StrSet
}

func NewCtgMq(ctx kctx.Ctx, getter IConfigGetter, prefix string, arg CtgMqCliPara) *CtgMq {
	if arg.WorkSize == 0 {
		arg.WorkSize = 50
	}
	if arg.MaxSpan == 0 {
		arg.MaxSpan = 1000
	}
	gen := conf.NewConfigKey(prefix)
	return &CtgMq{
		getter: getter,
		prefix: prefix,
		ctx:    ctx,
		arg:    arg,
		configList: gset.NewStrSetFrom([]string{
			gen.Key("Brokers"),
			gen.Key("UserName"),
			gen.Key("Password"),
			gen.Key("Group"),
		}),
	}
}

func (k *CtgMq) SetConfKeys(names []string) {
	k.configList = gset.NewStrSetFrom(names)
}

func (k *CtgMq) Consume(topicKey string, h model.HandleFunc) (cf func(), err error) {
	k.handleFunc = h

	if !k.getter.IsNacosSync() {
		// 不需要动态更新nacos 配置
		kafkaConf := NewProducerConfCtg(k.getter, k.prefix)
		kafkaConf.MessageModel = k.arg.MessageModel
		kafkaConf.Topic = k.getter.GetString(topicKey)
		cf, err = NewCtgMqCli(k.ctx, *kafkaConf, k.handleFunc, k.arg)
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
				ctgConf := NewProducerConfCtg(k.getter, k.prefix)
				if cf != nil {
					cf()
				}
				cf, err = NewCtgMqCli(k.ctx, *ctgConf, k.handleFunc, k.arg)
				if err != nil {
					k.ctx.Errorf("NewCtgConsumerCli failed: %v", err)
				}
			}
		}
	}()
	changed <- struct{}{}

	return
}

func (k *CtgMq) Producer() (sender ISender, cf func(), err error) {

	if !k.getter.IsNacosSync() {
		// 不需要动态更新nacos 配置
		cf, sender, err = NewRocketMqProducerWithConfig(k.ctx, k.getter, k.prefix)
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
				cf, sender, err = NewRocketMqProducerWithConfig(k.ctx, k.getter, k.prefix)
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

func NewProducerConfCtg(getter IConfigGetter, prefix string) *CtgMqCliConf {
	gen := conf.NewConfigKey(prefix)
	return &CtgMqCliConf{
		Brokers:  getter.GetString(gen.Key("Brokers")),
		Group:    getter.GetString(gen.Key("Group")),
		UserName: getter.GetString(gen.Key("UserName")),
		Password: getter.GetString(gen.Key("Password")),
	}
}
