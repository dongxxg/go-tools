## kit

     一个基于创研底座(unitechs.com/unios-dice/uni-base)进行二次封装的工具包。 

## Usage
 

### 1. 上下文的使用

kit 中 kctx 包里面包含了Ctx{}, 该结构体中封装了context.Context 和日志模块。 日志包含了trace_id。 建议ctx 从数据入口处一直传到业务逻辑结尾，
同一ctx 打印的trace_id 一样，方便追踪


```golang
package main

import (
	"github.com/spf13/viper"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
	ctx := kctx.NewDefaultContext()
	ctx.Info("这是一行日志")
	ctx.Infof("这也是%s", "一行日志")
	ctx.Error("这还是一行日志")
	ctx.Errorf("这又是%s", "一行日志")

	/*
		业务逻辑
	*/

	startup.Bootstrap(&startup.BootConfig{
		ServiceName: "demo",
		APIs:        nil,
	})

}
```

### 2. nacos 配置动态更新

1. 实现程序监听 nacos 配置变化并实现相关的操作。如配置变化涉及中间件的连接信息，则断开中间件，再重新创建连接。
   比如： nacos中mysql 的地址变了， 程序感知到地址变化后，关闭原来的连接并开启新的连接
2. 实现其他配置的动态监听
   如： nacos 配置变更，程序应最短时间内感知并应用

```golang 
package main

import (
	"github.com/spf13/viper"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {

	getter := conf.NewKitConf(
		viper.GetViper(),
		conf.WithNacosSync(true),   // 传入参数WithNacosSync=true  默认为false 
		conf.WithNacosGroup("Pub"), // 同步nacos配置变化到本地, 并传入组名
	)

	/*
		业务逻辑
	*/

	startup.Bootstrap(&startup.BootConfig{
		ServiceName: "demo",
		APIs:        nil,
	})

}

```

### 3. kafka

此kafka 生产者模型实现了动态监听nacos配置变化，如kafka 地址发生变化， 则关闭原连接，重新创建新的kafka连接。 该封装约定了配置文件的格式

**消息接收及处理模式**

开一个msgChannel 长度为 RecvSize * 2， 用于接收消息， 启动 workSize 个协程，处理msgChannel中的消息

#### 配置文件

```toml
[Topic]
Event = "test"

[Event.Kafka] # 接收告警信息
Brokers = "127.0.0.1:9092"
Oldest = false
Group = "demo"
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
KerberosAuthType = 2 # 1: KRB5_USER_AUTH  2: KRB5_KEYTAB_AUTH
KerberosConfigPath = ""
DisablePAFXFAST = false
```

#### 示例代码

 ```golang
 package main

import (
	"github.com/spf13/viper"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/client"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/cno/kit/util"
	"unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
	ctx := kctx.NewDefaultContext()
	getter := conf.NewKitConf(
		viper.GetViper(),
		conf.WithNacosSync(true),   // 传入参数WithNacosSync=true，使用nacos 配置文件同步;  默认为false，不使用 
		conf.WithNacosGroup("Pub"), // 同步nacos配置变化到本地, 并传入组名， 为空时: DEFAULT_GROUP
	)

	cli := client.NewKafka(ctx, getter, "Event.Kafka", client.KafkaArg{
		RecvSize: 1,
		WorkSize: 50,
	})
	defer cf()

	// 创建一个消费者， 去消费该topic
	cf2, err := cli.Consume("Topic.Event", func(c context.Context, msg model.ConsumerMsg) {
		ctx.Infof("in kafka msg: %s", string(msg.Value))
	})

	if err != nil {
		ctx.Fatal(err)
	}
	defer cf2()

	// 创建一个生产者
	sender, cf, err := cli.Producer()
	if err != nil {
		ctx.Fatal(err)
	}
	defer cf()

	// 循环生产消息
	for {
		time.Sleep(time.Second * 3)
		topic := getter.GetString("Topic.Event")
		if err := sender.Send(topic, util.NewUUIDStr(), "test content"); err != nil {
			ctx.Errorf("send to kafka %s failed: %v", topic, err)
		}
		ctx.Errorf("send  to kafka %s success", topic)
	}

	startup.Bootstrap(&startup.BootConfig{
		ServiceName: "demo",
		APIs:        nil,
	})
}
 ```

从启动到 nacos 配置变化，日志及说明如下:

```azure
time=2023-10-05T19:37:40+08:00 level=info msg=client/coordinator requesting coordinator for demo from
localhost:9092
time=2023-10-05
T19:37:40+08:00 level=info msg=client/coordinator coordinator for demo is #0 (localhost:9092)

// 1. 启动成功标志
time=2023-10-05T19:37:40+08:00 level=info msg=Sarama consumer up and running!...
time=2023-10-05T19:37:41+08:00 level=info msg=consumer/broker/0 accumulated 1 new subscriptions

// 2. 添加一个订阅到teset2
time=2023-10-05T19:37:41+08:00 level=info msg=consumer/broker/0 added subscription to test2/0

time=2023-10-05T19:37:43+08:00 level=info msg=Deprecation
warning: Consumer.Group.Rebalance.Strategy exists for historical compatibility and should not be used. Please
use Consumer.Group.Rebalance.GroupStrategies time=2023-10-05
T19:37:43+08:00 level=info msg=client/metadata fetching metadata for [test2] from broker
localhost:9092

time=2023-10-05
T19:37:43+08:00 level=info msg=Connected to broker at
localhost:9092 (unregistered)

time=2023-10-05
T19:37:43+08:00 level=info msg=client/brokers registered new broker #0 at
localhost:9092
time=2023-10-05
T19:37:43+08:00 level=info msg=Deprecation
warning: Consumer.Group.Rebalance.Strategy exists for historical compatibility and should not be used. Please
use Consumer.Group.Rebalance.GroupStrategies time=2023-10-05
T19:37:43+08:00 level=info msg=producer/broker/0 starting up

time=2023-10-05T19:37:43+08:00 level=info msg=producer/broker/0 state change to [open] on test2/0

time=2023-10-05T19:37:43+08:00 level=info msg=Connected to broker at
localhost:9092 (registered as #0)
    // 3. 发送消息到kafka test2 topic
    time =2023-10-05
T19:37:43+08:00 level=error msg=
send to kafka test2 success trace_id=676c86ba9de64fec9655673e06fa169d
// 4. 消费topic  test2 的消息  
time=2023-10-05
T19:37:43+08:00 level=info msg=in kafka
msg: test content trace_id=676c86ba9de64fec9655673e06fa169d
time=2023-10-05
T19:37:46+08:00 level=error msg=
send to kafka test2 success trace_id=676c86ba9de64fec9655673e06fa169d
time=2023-10-05
T19:37:46+08:00 level=info msg=in kafka
msg: test content trace_id=676c86ba9de64fec9655673e06fa169d
// 5. 关闭对 test2 topic 的订阅 
time=2023-10-05
T19:38:05+08:00 level=info msg=consumer/broker/0 closed dead subscription to test2/0

time=2023-10-05T19:38:05+08:00 level=info msg=consumergroup/session/unitechs_base_client-f4c5c882-d800-42d4-ace5-7c9f8d32b8ba/1 heartbeat loop stopped

time=2023-10-05T19:38:05+08:00 level=info msg=consumergroup/session/unitechs_base_client-f4c5c882-d800-42d4-ace5-7c9f8d32b8ba/1 released
// 6. 关闭kafka client 
time=2023-10-05T19:38:05+08:00 level=info msg=Closing Client
// 7. 新的kafka 的完成配置信息
time=2023-10-05T19:38:05+08:00 level=debug msg=kafka config {"Version":"2.0.0","Brokers":"localhost:9092","Topics":"test3",
"Group":"demo","Assignor":"range","Oldest":false,"SaslEnable":false,"UserName":"","Password":"","Algorithm":"",
"CertFile":"","CaFile":"","KeyFile":"","VerifySSL":false,"KerberosConfig":null,"Realm":"","ServiceName":"","KeyTabPath":"",
"KerberosConfigPath":"","UseTLS":false,"TokenProvider":null,
"Backoff":""} trace_id=676c86ba9de64fec9655673e06fa169d
time=2023-10-05
T19:38:05+08:00 level=info msg=Closed connection to broker
localhost:9092

time=2023-10-05
T19:38:05+08:00 level=info msg=Closed connection to broker
localhost:9092

time=2023-10-05
T19:38:05+08:00 level=info msg=Initializing new client

time=2023-10-05T19:38:05+08:00 level=info msg=client/coordinator coordinator for demo is #0 (localhost:9092)
// 8. 消费者起来， 开始消费
time=2023-10-05T19:38:05+08:00 level=info msg=Sarama consumer up and running!...
time=2023-10-05T19:38:05+08:00 level=info msg=consumer/broker/0 accumulated 1 new subscriptions
// 9. 添加对topic test3 的订阅
time=2023-10-05T19:38:05+08:00 level=info msg=consumer/broker/0 added subscription to test3/0

time=2023-10-05T19:38:07+08:00 level=info msg=client/metadata fetching metadata for [test3] from broker
localhost:9092

time=2023-10-05
T19:38:07+08:00 level=info msg=producer/broker/0 state change to [open] on test3/0
// 10. 发送消息到topic test3
time=2023-10-05T19:38:07+08:00 level=error msg=
send to kafka test3 success trace_id=676c86ba9de64fec9655673e06fa169d
// 11. 接收test3 topic 的消息 
time=2023-10-05
T19:38:07+08:00 level=info msg=in kafka
msg: test content trace_id=676c86ba9de64fec9655673e06fa169d
time=2023-10-05
T19:38:10+08:00 level=error msg=
send to kafka test3 success trace_id=676c86ba9de64fec9655673e06fa169d
time=2023-10-05
T19:38:10+08:00 level=info msg=in kafka msg: test content trace_id=676c86ba9de64fec9655673e06fa169d
```

### 4 ES 连接及使用方法

配置如下：

```toml
[Alarm.Es]
Urls = "http://localhost:9200"
# ES 类型 ， 6.几版本时， 需要配置, 7.几时，配置为空
Type = ""
Username = ""
Password = ""
# 生成入ES语句 等待PUB_ES_WAITTIME秒入单位秒， 如WatiTime为5时， 表示每5s， 如果有数据则入es一次
WaitTime = 5
# ${Pub.Es.Index.Prefix}am_ala_alarmsummary_${!now().format_timestamp("20060102", "${Alarm.Es.Zone}")}
Index = "am_ala_alarmsummary"
IndexLayout = "20060102"
```

实现 nacos 热更新
示例代码：

```golang

package main

import (
   "github.com/spf13/viper"
   "os"
   "time"
   "unitechs.com/cno/kit/kctx"
   "unitechs.com/cno/kit/service/client"
   "unitechs.com/cno/kit/service/conf"
   "unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
   ctx := kctx.NewDefaultContext()
   getter := conf.NewKitConf(viper.GetViper(), conf.WithNacosSync(true))
   os.Setenv("UNIBASE_CONFIG_PATH", "../../config")

   cli, err := client.NewEsRepoWith(ctx, getter, "Alarm.Es", "", func(ctx kctx.Ctx, infos []client.EsSendCap, esErr error, res map[string]interface{}) {
      /*

         发送异常信息的回调处理 逻辑... ...
         可以为空
      */
   })
   if err != nil {
      ctx.Fatal(err)
   }
   // 采用nacos热更新的模式，如果索引和时间后缀和默认的不一致的时候， 设置索引
   cli.SetIndexKey("Alarm.Es.Index") //  // 设置新的索引名称
   cli.SetIndexTimeLayoutKey("Alarm.Es.IndexLayout")
   if _, err := cli.Run(); err != nil {
      ctx.Fatal("run failed: %v", err)
   }

   info := map[string]string{
      "name": "法外狂徒",
      "age":  "18",
   }

   // send 传入id， 实现upsert
   if err = cli.Send(ctx, "id", info); err != nil {
      ctx.Errorf("send to es failed: %v", err)
   }

   go func() {
      for {
         // 循环打印配置，可以看到配置的热更新
         cfg := cli.GetCli().Config()
         ctx.Infof("host: %s, port: %s, username: %s, password: %s", cfg.Addresses, "", cfg.Username, cfg.Password)
         time.Sleep(time.Second)
      }
   }()

   startup.Bootstrap(&startup.BootConfig{
      ServiceName: "demo",
      APIs:        nil,
   })

}

```

### 5 kafka/ctg-mq 统一封装
该封装实现简单调用kafka及ctg-mq的消费者及生产者，是对上述kafka 及ctg-mq 的简单封装。所以配置文件与2，  3中一致
示例代码：
```golang
package main

import (
   "context"
   "github.com/spf13/viper"
   "os"
   "time"
   "unitechs.com/cno/kit/kctx"
   "unitechs.com/cno/kit/lib/mq/model"
   "unitechs.com/cno/kit/service/client"
   "unitechs.com/cno/kit/service/conf"
   "unitechs.com/cno/kit/util"
   "unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
   ctx := kctx.NewDefaultContext()
   getter := conf.NewKitConf(viper.GetViper(), conf.WithNacosSync(true))
   os.Setenv("UNIBASE_CONFIG_PATH", "../../config")

   // 初始化 生产者， 这里发送模式kafka , 可根据程序不同的标记位，传入不同的sendType
   cf, sender, err := client.NewProducerCli(ctx, client.SendTypeKafka, getter, "Event.kafka", "")
   if err != nil {
      ctx.Fatal(err)
   }
   defer cf()
   go func() {
      for {
		  // 循环写入内容到蹭件
         time.Sleep(time.Second)
         topic := getter.GetString("Topic.Event")
         if err_ := sender.Send(topic, util.NewUUIDStr(), "test content"); err_ != nil {
            ctx.Info("send to kafka failed: %v", err)
            continue
         }
         ctx.Infof("send to kafka %s success %s", topic, "test content")
      }
   }()
   
   
    // 初始化消费者 并注入消费 函数 
   cf2, err := client.NewConsumerCliWithNacos(ctx, client.SendTypeKafka, getter, func(c context.Context, msg model.ConsumerMsg) {
      ctx.Infof("in consumer %s", string(msg.Value))
   }, client.CliArg{
      TopicKey:       "Topic.Event",  // 需要注意，这里是topic的配置文件的key值， 并非topic值本身
      WorkSize:       10,
      KafkaPrefix:    "Event.Kafka", 
      RocketmqPrefix: "Event.Rocketmq",
      RocketmqMsgMod: "", 
   })
   if err != nil {
      ctx.Fatal(err)
   }
   defer cf2()
   
   startup.Bootstrap(&startup.BootConfig{
      ServiceName: "demo",
      APIs:        nil,
   })
}

```

### 6. notify 通知实现
```golang 
package main

import (
   "github.com/spf13/viper"
   "os"
   "time"
   "unitechs.com/cno/kit/kctx"
   "unitechs.com/cno/kit/service/client"
   "unitechs.com/cno/kit/service/conf"
   "unitechs.com/cno/kit/util"
   "unitechs.com/unios-dice/uni-base/core/startup"
)

func main() {
   ctx := kctx.NewDefaultContext()
   getter := conf.NewKitConf(viper.GetViper(), conf.WithNacosSync(true))
   os.Setenv("UNIBASE_CONFIG_PATH", "../../config")

   cf, err := client.NewNotifyCli(ctx, client.MiddlewareTypeKafka, getter, client.CliArg{
      TopicKey:       "Topic.Notify",
      WorkSize:       10,
      MaxSpan:        2000,
      KafkaPrefix:    "Notify.Kafka",
      RocketmqPrefix: "Notify.Rocketmq",
      RocketmqMsgMod: "BroadCasting", // ctg-mq开启 广播模式
   }).Registry("507", func(topic string, startTime time.Time, value []byte) {
      ctx.Infof("consumer msg 507 %s", string(value))
   }).Registry("503", func(topic string, startTime time.Time, value []byte) {
      ctx.Infof("consumer msg 503 %s", string(value))
   }).Run()
   if err != nil {
      ctx.Fatal(err)
   }
   defer cf()
   
    // 新生成一个生产者， 模拟发消息
   cli2 := client.NewKafka(ctx, getter, "Event.Kafka", client.KafkaArg{
      WorkSize: 10,
   })
   sender, cf2, err := cli2.Producer()
   if err != nil {
      ctx.Fatal(err)
   }
   go func() {
      for {
         time.Sleep(time.Second)
         topic := getter.GetString("Topic.Notify")
         if err_ := sender.Send(topic, util.NewUUIDStr(), "{\"msg_head\": {\"type\": \"507\"}}"); err_ != nil {
            ctx.Info("send to kafka failed: %v", err)
            continue
         }
         ctx.Infof("send to kafka %s success %s", topic, "test content")
      }
   }()
   defer cf2()

   startup.Bootstrap(&startup.BootConfig{
      ServiceName: "demo",
      APIs:        nil,
   })
}
```
日志如下：
```azure
time=2023-10-06T15:44:49+08:00 level=info msg=Connected to broker at localhost:9092 (unregistered)

time=2023-10-06T15:44:49+08:00 level=info msg=client/brokers registered new broker #0 at localhost:9092
time=2023-10-06T15:44:49+08:00 level=info msg=Deprecation warning: Consumer.Group.Rebalance.Strategy exists for historical compatibility and should not be used. Please use Consumer.Group.Rebalance.GroupStrategies
time=2023-10-06T15:44:49+08:00 level=info msg=producer/broker/0 starting up

time=2023-10-06T15:44:49+08:00 level=info msg=producer/broker/0 state change to [open] on test/0

time=2023-10-06T15:44:49+08:00 level=info msg=Connected to broker at localhost:9092 (registered as #0)

time=2023-10-06T15:44:49+08:00 level=info msg=send to kafka test success test content trace_id=831879e026614928898c4a80527a5404
time=2023-10-06T15:44:49+08:00 level=info msg=input data notify  {"msg_head": {"type": "507"}} trace_id=458ea276d07b445ba7d294cc13ccb692
time=2023-10-06T15:44:49+08:00 level=info msg=consumer msg 507 {"msg_head": {"type": "507"}} trace_id=831879e026614928898c4a80527a5404
time=2023-10-06T15:44:49+08:00 level=info msg=deal notify 507 success trace_id=458ea276d07b445ba7d294cc13ccb692
time=2023-10-06T15:44:50+08:00 level=info msg=send to kafka test success test content trace_id=831879e026614928898c4a80527a5404
time=2023-10-06T15:44:50+08:00 level=info msg=input data notify  {"msg_head": {"type": "507"}} trace_id=c0f1b64cbb1b43bb8cb63752d5e9ec5d
time=2023-10-06T15:44:50+08:00 level=info msg=consumer msg 507 {"msg_head": {"type": "507"}} trace_id=831879e026614928898c4a80527a5404
time=2023-10-06T15:44:50+08:00 level=info msg=deal notify 507 success trace_id=c0f1b64cbb1b43bb8cb63752d5e9ec5d

```

    