package client

import (
	"context"
	"fmt"
	"testing"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/service/conf"
)

func TestNotify(t *testing.T) {
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaa")
	go notifyDemo(ctx, "demo1", MiddlewareTypeKafka, "")
	go notifyDemo(ctx, "demo2", MiddlewareTypeKafka, "")
	select {}
}

func TestNotifyRocketmq(t *testing.T) {
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaa")

	notifyDemo(ctx, "demo1", MiddlewareTypeRocketmq, "BroadCasting")
	//go notifyDemo(ctx, "demo2", MiddlewareTypeRocketmq, "BroadCasting")
	//go notifyDemo(ctx, "demo3", MiddlewareTypeRocketmq, "")
	select {}
}

func notifyDemo(ctx kctx.Ctx, name string, m MiddlewareType, msgMod string) {
	getter := conf.NewTestKitConf("../../config/")

	cf, err := NewNotifyCli(ctx, m, getter, CliArg{
		Topic:          "unios-alarm-notify",
		RecvSize:       1,
		WorkSize:       1,
		MaxSpan:        1000,
		KafkaPrefix:    "Event.Kafka",
		RocketmqPrefix: "Event.Rocketmq",
		RocketmqMsgMod: msgMod,
	}).Registry("507", func(topic string, startTime time.Time, value []byte) {
		fmt.Println(name, "in registry1", string(value))
	}).Registry("302", func(topic string, startTime time.Time, value []byte) {
		fmt.Println(name, "in registry2", string(value))
	}).Run()
	if err != nil {
		ctx.Info(err)
	}
	defer cf()
	ctx.Info("over")
	select {}
}
