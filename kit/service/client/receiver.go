package client

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	"unitechs.com/cno/kit/lib/mq/model"
)

const (
	RecvTypeKafka = iota + 1
	RecvTypeCtgMq
)

type RecvNotify struct {
	Type      int
	kafkaConf kafka.Config
	ctgmqConf []consumer.Option
	ctx       kctx.Ctx
}
type CliArg struct {
	IsGroupFuncOpen bool   // // 是否打开 消费组生成函数
	TopicKey        string // topic在配置文件上的key，如Topic.Event， 注意不是nacos的TOPIC_EVENT
	Topic           string
	RecvSize        int                                           // 容量
	WorkSize        int                                           // 工作线程数量
	MaxSpan         int                                           // ctg-mq 使用
	KafkaPrefix     string                                        // 配置文件前缀
	RocketmqPrefix  string                                        // 配置文件前缀
	RocketmqMsgMod  string                                        //  消息模式  BroadCasting/Clustering
	KafkaGroup      func(group, nacosPort string) (string, error) // 消费组有不同的要求的时候, 比如容器环境下 notify的时候， 要求每个实例不一样
	Group           string                                        // 指定消费组
	MQGroup         func(group, nacosPort string) (string, error)
}

func NewCli(ctx kctx.Ctx, RecvType int, kConf *kafka.Config, conf *CtgMqCliConf,
	handFunc model.HandleFunc, arg CliArg) (
	cf func(), err error) {
	switch RecvType {
	case RecvTypeKafka:
		if kConf.Algorithm == "none" {
			kConf.Algorithm = ""
		}
		ctx.Info("recv type is kafka")
		return NewKafkaConsumerCli(ctx, kConf, handFunc, KafkaArg{
			RecvSize:  arg.RecvSize,
			WorkSize:  arg.WorkSize,
			GroupFunc: arg.KafkaGroup,
		})
	case RecvTypeCtgMq:
		ctx.Info("recv type is ctg-mq")
		return NewCtgMqCli(ctx, *conf, handFunc, CtgMqCliPara{
			WorkSize: arg.WorkSize,
			MaxSpan:  arg.MaxSpan,
		})
	default:
		return func() {

		}, fmt.Errorf("not supported RecvType %d", RecvType)
	}
}

// NewConsumerCli
/*
# kafka 的配置
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

[Rocketmq]
Brokers = "127.0.0.1:9876"
Group = "unios-alarm-process-notify-group" # 注意，千万不能为空
Username = ""
Password = ""
# 实例名称， 非必填
InstanceName = ""
*/
func NewConsumerCli(ctx kctx.Ctx, consumerType int, getter IConfigGetter, f model.HandleFunc,
	cliArg CliArg) (func(), error) {
	switch consumerType {
	case RecvTypeKafka:
		return NewCli(ctx, RecvTypeKafka, &kafka.Config{
			Version: getter.GetString(fmt.Sprintf("%s.Version", cliArg.KafkaPrefix)),
			Brokers: getter.GetString(fmt.Sprintf("%s.Brokers", cliArg.KafkaPrefix)),
			Topics:  cliArg.Topic,
			Group: func() string {
				if cliArg.Group != "" {
					return cliArg.Group
				}
				return getter.GetString(fmt.Sprintf("%s.Group", cliArg.KafkaPrefix))
			}(),
			Assignor:                     getter.GetString(fmt.Sprintf("%s.Assignor", cliArg.KafkaPrefix)),
			Oldest:                       getter.GetBool(fmt.Sprintf("%s.Oldest", cliArg.KafkaPrefix)),
			SaslEnable:                   getter.GetBool(fmt.Sprintf("%s.SaslEnable", cliArg.KafkaPrefix)),
			UserName:                     getter.GetString(fmt.Sprintf("%s.UserName", cliArg.KafkaPrefix)),
			Password:                     getter.GetString(fmt.Sprintf("%s.Password", cliArg.KafkaPrefix)),
			Algorithm:                    getter.GetString(fmt.Sprintf("%s.Algorithm", cliArg.KafkaPrefix)),
			CertFile:                     getter.GetString(fmt.Sprintf("%s.CertFile", cliArg.KafkaPrefix)),
			CaFile:                       getter.GetString(fmt.Sprintf("%s.CaFile", cliArg.KafkaPrefix)),
			KeyFile:                      getter.GetString(fmt.Sprintf("%s.KeyFile", cliArg.KafkaPrefix)),
			VerifySSL:                    getter.GetBool(fmt.Sprintf("%s.VerifySSL", cliArg.KafkaPrefix)),
			Realm:                        getter.GetString(fmt.Sprintf("%s.Realm", cliArg.KafkaPrefix)),
			ServiceName:                  getter.GetString(fmt.Sprintf("%s.ServiceName", cliArg.KafkaPrefix)),
			KeyTabPath:                   getter.GetString(fmt.Sprintf("%s.KeyTabPath", cliArg.KafkaPrefix)),
			KerberosConfigPath:           getter.GetString(fmt.Sprintf("%s.KerberosConfigPath", cliArg.KafkaPrefix)),
			Backoff:                      getter.GetString(fmt.Sprintf("%s.Backoff", cliArg.KafkaPrefix)),
			MaxProcessingTimeMilliSecond: getter.GetInt(fmt.Sprintf("%s.MaxProcessingTimeMilliSecond", cliArg.KafkaPrefix)),
		}, nil, f, cliArg)

	case RecvTypeCtgMq:
		return NewCli(ctx, RecvTypeCtgMq, nil, &CtgMqCliConf{
			Brokers:  getter.GetString(fmt.Sprintf("%s.Brokers", cliArg.RocketmqPrefix)),
			Topic:    cliArg.Topic,
			UserName: getter.GetString(fmt.Sprintf("%s.UserName", cliArg.RocketmqPrefix)),
			Password: getter.GetString(fmt.Sprintf("%s.Password", cliArg.RocketmqPrefix)),
			Group: func() string {
				if cliArg.Group != "" {
					return cliArg.Group
				}
				return getter.GetString(fmt.Sprintf("%s.Group", cliArg.RocketmqPrefix))
			}(),
			MessageModel: cliArg.RocketmqMsgMod,
		}, f, cliArg)
	default:
		return func() {}, fmt.Errorf("invalid send mod %d,  1: kafka 2:ctg-mq", consumerType)
	}
}

// NewConsumerCliWithNacos 生成消费者client, 根据传入消费类型，自动消费， 根据传入类型是kafka/ctg-mq 对特定的topic 进行消费
// 本方法支持 nacos 热更新
func NewConsumerCliWithNacos(ctx kctx.Ctx, t MiddlewareType, getter IConfigGetter, h model.HandleFunc, arg *CliArg) (func(), error) {
	switch t {
	case MiddlewareTypeKafka:
		cli := NewKafka(ctx, getter, arg.KafkaPrefix, KafkaArg{
			WorkSize:        arg.WorkSize,
			GroupFunc:       arg.KafkaGroup,
			IsGroupFuncOpen: arg.IsGroupFuncOpen,
		})
		return cli.Consume(arg.TopicKey, h)

	case MiddlewareTypeRocketmq:
		cli := NewCtgMq(ctx, getter, arg.RocketmqPrefix, CtgMqCliPara{
			WorkSize:     arg.WorkSize,
			MaxSpan:      arg.MaxSpan,
			MessageModel: arg.RocketmqMsgMod,
		})
		return cli.Consume(arg.TopicKey, h)
	default:
		return func() {}, fmt.Errorf("invalid send mod %s", t)
	}
}
