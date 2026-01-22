package client

import (
	"context"
	"testing"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/cno/kit/util"
)

func TestProducer(t *testing.T) {
	ctx := kctx.NewContext(context.TODO(), "cccccccccc")

	getter := conf.NewTestKitConf("../../config/")
	cf, sender, err := NewProducerCli(ctx, "kafka", getter, "Event.Kafka", "Event.Rocketmq")
	if err != nil {
		ctx.Fatal(err)
	}
	defer cf()

	for i := 0; i < 10; i++ {
		if err = sender.Send("unios-alarm-notify", util.NewUUIDStr(), `{"msg_head":{"type":"302"}}`); err != nil {
			ctx.Errorf(err.Error())
		}
		ctx.Info("succss")
	}
}
