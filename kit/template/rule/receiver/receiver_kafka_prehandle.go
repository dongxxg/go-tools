package receiver

import (
	"context"
	"github.com/pkg/errors"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/template/cache"
	model2 "unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
)

/*
连接RECV_KAFKA，按对应topic实时接收数据流。
若消息中type=AlarmStatusChange.AlarmRelation，则遍历msg中subAlarms中alarmUniqueClearId字段，调用http://PUB_ALARM_SERVICE/api/ala/prehandle/execByAlarmId接口，其中入参sumAlarmId=alarmUniqueClearId。否则，丢弃消息。
若返回code=0000，则打印调用成功信息，否则打印失败信息。


消息样例;
{
	"encoding": "GBK",
	"msg": "{\"alarmProvince\":\"吉林省\",\"alarmUniqueClearId\":\"3612132270_1999365390_2808922174_2657692148\",\"alarmUniqueId\":\"3612132270_1999365390_2808922174_2657692148\",\"eventTime\":\"2022-03-12 12:10:56\",\"subAlarms\":[{\"alarmProvince\":\"吉林省\",\"alarmUniqueClearId\":\"E3E5BFCBDE9B9C32E05354801DAC0C65\",\"alarmUniqueId\":\"E8E9BFCBDE9B9C32E05354801DAC0C63\",\"eventTime\":\"2022-03-12 12:12:56\"},{\"alarmProvince\":\"吉林省\",\"alarmUniqueClearId\":\"E7E5BFCBDE9B9C32E05354801DAC0C60\",\"alarmUniqueId\":\"E8E5BFCBDE9B9C32E05354801DAC0C61\",\"eventTime\":\"2022-03-12 12:10:56\"}]}",
	"msgSerial": "1637573111999",
	"systemName": "FMC",
	"type": "AlarmStatusChange.AlarmRelation"
}
*/

type ReceiverPrehandle struct {
	p          model2.ParamKafkaPrehandle
	c          *kafka.Consumer
	ctx        kctx.Ctx
	cancelFunc func() error
}

func NewReceiverPrehandle(ctx kctx.Ctx, p model2.ParamKafkaPrehandle) (*ReceiverPrehandle, error) {
	c, err := kafka.NewConsumerWith(kafka.ParamKafka{
		Broker:    p.RecvKafkaPrehandleBrokers,
		Topic:     p.RecvKafkaPrehandleTopic,
		User:      p.RecvKafkaPrehandleSaslUsername,
		Pass:      p.RecvKafkaPrehandleSaslPassword,
		Algorithm: kafka.Algorithm(p.RecvKafkaPrehandleSaslAlgorithm),
		Group:     p.RecvKafkaPrehandleGroup,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "new kafka consumer failed")
	}

	return &ReceiverPrehandle{
		p:   p,
		c:   c,
		ctx: ctx,
	}, nil
}

func (r *ReceiverPrehandle) Collect(ruleId string, param map[string]interface{}, f model.HandleFunc) (cf func(), err error) {
	//ctx_, cf_ := context.WithCancel(r.ctx.Context)
	//r.cf = cf_

	p := model2.ParamKafkaPrehandle{}
	if err = util.MarshalWith(param, &p); err != nil {
		err = errors.Wrapf(err, "failed to parse prehadle  param: %+v", param)
		return
	}
	r.ctx.Debugf("param %+v", p)
	if err = util.MarshalWith(param, &p); err != nil {
		r.ctx.Errorf("failed to parse prehandle param, err: %v, param: %+v", err, param)
		return
	}

	p.RecvKafkaPrehandleSaslPassword = cache.GetPassword(r.ctx, p.RecvKafkaPrehandleSaslPassword, "RECV_KAFKA_PREHANDLE_SASL_PASSWORD")

	ctx := context.WithValue(context.TODO(), KeyRuleID, ruleId)
	ctx = context.WithValue(context.TODO(), KeyRuleType, model2.RuleTypeKafkaPrehandle)
	ctx = context.WithValue(ctx, "param", r.p)
	c2 := kctx.NewContext(ctx, util.NewUUIDStr())
	cf2, err := r.c.Consume(c2, f)

	r.cancelFunc = func() error {
		return cf2()
	}
	return func() {
		r.Close()
	}, err
}
func (r *ReceiverPrehandle) Close() error {
	return r.cancelFunc()
}
