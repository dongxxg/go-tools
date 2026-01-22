package client

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/util"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

// notify 通知消息
var notify *NotifyCli

func GetNotify() *NotifyCli {
	return notify
}

type NotifyCli struct {
	rds        *redis.Redis
	data       sync.Map
	getter     IConfigGetter
	ctx        kctx.Ctx
	MiddleType MiddlewareType // 消费类型 1:kafka  2: rocketmq
	CliArg     *CliArg
}

// NewNotifyCli
/*
[xxx]
Brokers = "127.0.0.1:9092"
Oldest = false
Group = ""
Version = "2.0.0"

# 底坐客户端连接使用的 sha512 sha256 gssapi plain
#KafkaAlgorithm = "sha512"

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

[yyyy]
Brokers = "127.0.0.1:9876"
Username = ""
Password = ""
Group = "unios-alarm-process-group"
# 流控， 默认1000
ConsumerMaxSpan = 2000
*/
func NewNotifyCli(ctx kctx.Ctx, m MiddlewareType, getter IConfigGetter, arg CliArg) *NotifyCli {
	return &NotifyCli{
		getter:     getter,
		ctx:        ctx,
		MiddleType: m,
		CliArg:     &arg,
	}
}

type NotifyHandler func(topic string, startTime time.Time, value []byte)

func (n *NotifyCli) Registry(notifyCode string, f NotifyHandler) *NotifyCli {
	n.data.Store(notifyCode, f)
	return n
}

func (n *NotifyCli) RegistryWith(notifyCode string, f model.HandleFunc) *NotifyCli {
	n.data.Store(notifyCode, f)
	return n
}
func (n *NotifyCli) SetRedis(rds *redis.Redis) {
	n.rds = rds
}

func (n *NotifyCli) Run() (cf func(), err error) {
	if n.CliArg.KafkaGroup == nil {
		// 没有传入group 生成方法， 使用默认的
		n.CliArg.KafkaGroup = GenNotifyGroup
	}
	cf, err = NewConsumerCliWithNacos(n.ctx, n.MiddleType, n.getter, n.handler, n.CliArg)
	if err != nil {
		err = errors.Wrapf(err, "NotifyCli.Run")
	}

	return
}

// GenNotifyGroup 生成Notify的消费组
func GenNotifyGroup(group, nacosPort string) (string, error) {
	// 获取当前ip
	ip, err := util.GetLocalIP()
	if err != nil {
		return "", errors.Wrapf(err, "get local ip")
	}
	g := fmt.Sprintf("%s_%s_%s", group, ip.String(), nacosPort)

	return g, nil
}

//	type Head struct {
//		Client  string `json:"client"`  //发送者
//		Type    string `json:"type"`    //消息类型 501 502 503 504
//		SubType string `json:"subtype"` //消息子类型 add/mod/del
//		TraceId string `json:"trace_id"`
//	}
//
//	type Message struct {
//		MsgHead Head `json:"msg_head"` //如果是高级给你消息则为100
//		//MsgBody interface{} `json:"msg_body"`
//	}
func (n *NotifyCli) handler(c context.Context, msg model.ConsumerMsg) {
	g := gjson.ParseBytes(msg.Value)
	notifyCoe := g.Get("msg_head.type").String()
	if notifyCoe == "" {
		n.ctx.Errorf("invalid notify msg, msg_head.type not exist， %s", string(msg.Value))
		return
	}
	tid := g.Get("msg_head.trace_id").String()
	ctx := kctx.NewContext(c, tid)
	ctx.Info("input data notify  ", string(msg.Value))
	if f, ok := n.data.Load(notifyCoe); ok {
		if f1, ok := f.(NotifyHandler); ok {
			f1(msg.Topic, msg.Timestamp, msg.Value)
			ctx.Infof("deal notify %s success", notifyCoe)

		} else if f2, ok := f.(model.HandleFunc); ok {
			f2(context.TODO(), msg)
			ctx.Infof("deal notify %s success", notifyCoe)

		} else {
			n.data.Delete(notifyCoe)
		}
	} else {
		ctx.Errorf("invalid notifyCode %s", notifyCoe)
	}
}
