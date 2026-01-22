package main

import (
	"github.com/spf13/viper"
	"os"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
	ctx := kctx.NewDefaultContext()
	getter := conf.NewKitConf(viper.GetViper(), conf.WithNacosSync(true))
	os.Setenv("UNIBASE_CONFIG_PATH", "../../config")

	add := getter.GetString("Pub.ES.Urls")
	addr := getter.GetString("Pub.ES.Urls")
	ctx.Info(addr)
	ctx.Info(add)

	//db, dbType, err := client.NewDBCliWithNacos(ctx, getter, "Mysql")
	//if err != nil {
	//	ctx.Fatal()
	//}
	//ctx.Infof("dbType %s", dbType)
	//defer db.Close()
	//
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		rows, err := db.Query("select count(*) from am_ala_active_alarm")
	//		if err != nil {
	//			ctx.Error(err)
	//		}
	//		for rows.Next() {
	//			var n int
	//			rows.Scan(&n)
	//			ctx.Infof("count %d", n)
	//		}
	//	}
	//}()

	// notify
	//cf, err := client.NewNotifyCli(ctx, client.MiddlewareTypeKafka, getter, client.CliArg{
	//	TopicKey:       "Topic.Notify",
	//	WorkSize:       10,
	//	MaxSpan:        2000,
	//	KafkaPrefix:    "Notify.Kafka",
	//	RocketmqPrefix: "Notify.Rocketmq",
	//	RocketmqMsgMod: "BroadCasting",
	//}).Registry("507", func(topic string, startTime time.Time, value []byte) {
	//	ctx.Infof("consumer msg 507 %s", string(value))
	//}).Registry("503", func(topic string, startTime time.Time, value []byte) {
	//	ctx.Infof("consumer msg 503 %s", string(value))
	//}).Run()
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf()
	//
	//cli2 := client.NewKafka(ctx, getter, "Event.Kafka", client.KafkaArg{
	//	WorkSize: 10,
	//})
	//sender, cf2, err := cli2.Producer()
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		topic := getter.GetString("Topic.Notify")
	//		if err_ := sender.Send(topic, util.NewUUIDStr(), "{\"msg_head\": {\"type\": \"507\"}}"); err_ != nil {
	//			ctx.Info("send to kafka failed: %v", err)
	//			continue
	//		}
	//		ctx.Infof("send to kafka %s success %s", topic, "test content")
	//	}
	//}()
	//defer cf2()

	// kafka/ctg-mq
	//cf, sender, err := client.NewProducerCli(ctx, client.MiddlewareTypeKafka, getter, "Event.kafka", "")
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf()
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		topic := getter.GetString("Topic.Event")
	//		if err_ := sender.Send(topic, util.NewUUIDStr(), "test content"); err_ != nil {
	//			ctx.Info("send to kafka failed: %v", err)
	//			continue
	//		}
	//		ctx.Infof("send to kafka %s success %s", topic, "test content")
	//	}
	//}()
	//
	//cf2, err := client.NewConsumerCliWithNacos(ctx, client.MiddlewareTypeKafka, getter, func(c context.Context, msg model.ConsumerMsg) {
	//	ctx.Infof("in consumer %s", string(msg.Value))
	//}, client.CliArg{
	//	TopicKey:       "Topic.Event",
	//	WorkSize:       10,
	//	KafkaPrefix:    "Event.Kafka",
	//	RocketmqPrefix: "Event.Rocketmq",
	//	RocketmqMsgMod: "",
	//})
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf2()

	// ctg-cache
	//mq := client.NewCtgMq(ctx, getter, "Event.Rocketmq", client.CtgMqCliPara{
	//	WorkSize: 50,
	//	MaxSpan:  2000,
	//})
	//
	//topicKey := "Topic.Event"
	//cf, err := mq.Consume(topicKey, func(c context.Context, msg model.ConsumerMsg) {
	//	ctx.Infof("consume info %s", string(msg.Value))
	//})
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf()
	//
	//sender, cf2, err := mq.Producer()
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf2()
	//
	//go func() {
	//	for {
	//		topic := getter.GetString(topicKey)
	//		if err := sender.Send(topic, util.NewUUIDStr(), "test content"); err != nil {
	//			ctx.Errorf("send to ctg-mq failed: %v", err)
	//			continue
	//		}
	//		ctx.Infof("send to ctg-mq success topic: %s", topic)
	//	}
	//}()

	//cli, err := client.NewEsRepoWith(ctx, getter, "Pub.ES", make(chan client.EsSendCap), func(ctx kctx.Ctx, infos []client.EsSendCap, esErr error, res map[string]interface{}) {
	//	ctx.Errorf("es failed： %v", esErr)
	//})
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//// 采用nacos热更新的模式，如果索引和时间后缀和默认的不一致的时候， 设置索引
	//cli.SetIndexKey("Alarm.Es.Index") //  // 设置新的索引名称
	//cli.SetIndexTimeLayoutKey("Alarm.Es.IndexLayout")
	//if _, err := cli.Run(); err != nil {
	//	ctx.Fatal("run failed: %v", err)
	//}
	//info := map[string]string{
	//	"name": "法外狂徒",
	//	"age":  "18",
	//}
	//
	//// send 传入id， 实现upsert
	//go func() {
	//	for {
	//		if err = cli.Send(ctx, "hhhh", info); err != nil {
	//			ctx.Errorf("send to es failed: %v", err)
	//			continue
	//		}
	//		ctx.Infof("send to es success")
	//		time.Sleep(time.Second * 3)
	//	}
	//}()
	//
	//go func() {
	//	for {
	//		cfg := cli.GetCli().Config()
	//		ctx.Infof("host: %s, port: %s, username: %s, password: %s", cfg.Addresses, "", cfg.Username, cfg.Password)
	//		time.Sleep(time.Second * 4)
	//	}
	//}()

	// redis
	//conn := client.NewRedisCli(ctx, getter, "Redis")
	//
	//go func() {
	//	for {
	//		ctx.Infof("%s %s %d %s", conn.Type, conn.Addr, conn.Db, conn.Pass)
	//		time.Sleep(time.Second * 1)
	//	}
	//}()

	// kafka
	//cli := client.NewKafka(ctx, getter, "Event.Kafka", client.KafkaArg{
	//	RecvSize: 1,
	//	WorkSize: 50,
	//})
	//sender, cf, err := cli.Producer()
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf()
	//
	//cf2, err := cli.Consume("Topic.Event", func(c context.Context, msg model.ConsumerMsg) {
	//	ctx.Infof("in kafka msg: %s", string(msg.Value))
	//})
	//if err != nil {
	//	ctx.Fatal(err)
	//}
	//defer cf2()
	//
	//for {
	//	time.Sleep(time.Second * 3)
	//	topic := getter.GetString("Topic.Event")
	//	if err := sender.Send(topic, util.NewUUIDStr(), "test content"); err != nil {
	//		ctx.Errorf("send to kafka %s failed: %v", topic, err)
	//	}
	//	ctx.Errorf("send  to kafka %s success", topic)
	//}

	startup.Bootstrap(&startup.BootConfig{
		ServiceName: "demo",
		APIs:        nil,
	})

}
