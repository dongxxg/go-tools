package receiver

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/pkg/errors"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/client"
	model2 "unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
)

type ReceiverCTGMQ struct {
	ctx                 kctx.Ctx
	RetryIntervalSecond int
	consume             rocketmq.PushConsumer
	closeFunc           func()
	cancelFunc          context.CancelFunc
	PoolSize            int // 工作线程数据
}

func NewReceiverCTGMQ(ctx kctx.Ctx, topic string, options []consumer.Option, retryIntervalSecond int) *ReceiverCTGMQ {
	if retryIntervalSecond == 0 {
		retryIntervalSecond = 10
	}
	return &ReceiverCTGMQ{
		ctx:                 ctx,
		RetryIntervalSecond: retryIntervalSecond,
	}
}

func NewReceiverCTGMQWith(ctx kctx.Ctx, retryIntervalSecond int) (*ReceiverCTGMQ, error) {
	//if group == "" {
	//	return nil, errors.New("group of ctg-mq can`t be empty ")
	//}
	//options := make([]consumer.Option, 0)
	//
	//options = append(options, consumer.WithNameServer(endPoint))
	//options = append(options, consumer.WithGroupName(group))
	//if user != "" && pwd != "" {
	//	options = append(options, consumer.WithCredentials(primitive.Credentials{
	//		AccessKey: user,
	//		SecretKey: pwd,
	//	}))
	//}
	if retryIntervalSecond == 0 {
		retryIntervalSecond = 10
	}

	return &ReceiverCTGMQ{
		//options:             options,
		ctx:                 ctx,
		RetryIntervalSecond: retryIntervalSecond,
		//topic:               topic,
	}, nil
}
func (c *ReceiverCTGMQ) Close() error {
	return c.consume.Shutdown()
}
func (c *ReceiverCTGMQ) Collect(ruleId string, param map[string]interface{}, f model.HandleFunc) (cf func(), err error) {

	p := model2.ParamCTGMQ{}
	if err = util.MarshalWith(param, &p); err != nil {
		err = errors.Wrapf(err, "failed to parse ctg-mq param,param: %#v", param)
		return
	}
	conf := client.CtgMqCliConf{
		Brokers:  p.CTGmqBroker,
		Topic:    p.CTGmqTopic,
		UserName: p.CTGmqUser,
		Password: p.CTGmqPass,
		Group:    p.CTGmqGroup,
	}
	if c.PoolSize == 0 {
		c.PoolSize = 100
	}
	c.ctx.Context = context.WithValue(c.ctx, KeyRuleID, ruleId)

	cf, err = client.NewCtgMqCli(c.ctx, conf, f, client.CtgMqCliPara{
		WorkSize: c.PoolSize,
		MaxSpan:  2000,
	})
	c.cancelFunc = cf

	return
}
