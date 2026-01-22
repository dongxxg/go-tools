package main

import (
	"context"
	"github.com/spf13/viper"
	"unitechs.com/cno/kit/component-base/log2kafka"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/client"
	"unitechs.com/cno/kit/service/conf"
)

func main() {
	ctx := log2kafka.NewDefaultContext()
	ctx.Infof("begin")
	getter := conf.NewKitConf(viper.GetViper())
	cf := client.NewConsumerKafkaConf(getter, getter.GetString("Topic.Event"), "Event.Kafka")
	c, err := client.NewKafkaConsumerCli(ctx, cf, func(c context.Context, msg model.ConsumerMsg) {
		ctx.Infof(string(msg.Value))
	}, client.KafkaArg{
		RecvSize: 1,
		WorkSize: 1,
	})
	if err != nil {
		ctx.Error(err)
	}
	defer c()
	select {}
}
