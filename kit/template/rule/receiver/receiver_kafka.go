package receiver

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/mq/kafka"
	model2 "unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/service/client"
	"unitechs.com/cno/kit/template/cache"
	"unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
)

type ReceiverKafka struct {
	//conf                *kafka.Config
	RetryIntervalSecond int64
	ctx                 kctx.Ctx
	consumer            *kafka.Consumer
	cancelFunc          func() error
	cf                  context.CancelFunc
	WorkSize            int
}

func NewReceiverKafka(ctx kctx.Ctx, workSize int) (*ReceiverKafka, error) {
	return &ReceiverKafka{
		WorkSize:            workSize,
		ctx:                 ctx,
		RetryIntervalSecond: 5,
		//conf: conf,
	}, nil
}

const (
	keyTab       = "keytab"
	kerberosConf = "kbsConf"
)

func genKafkaConf(ctx kctx.Ctx, subPara map[string]interface{}) (*kafka.Config, error) {
	p := model.ParamKafka{}
	if err := util.UnMarshalWithInterface(subPara, &p); err != nil {
		return nil, errors.Wrapf(err, "invalid kafka subPara %#v", subPara)
	}

	c := kafka.NewConsumerConfig(
		p.KafkaBroker,
		p.KafkaTopic,
		false,
	)
	if p.KafkaVersion == "" {
		c.Version = "2.0.0"
	} else {
		c.Version = p.KafkaVersion
	}

	c.Group = p.KafkaGroup
	if p.KafkaUser != "" || p.KafkaPass != "" {
		c.SaslEnable = true
		c.UserName = p.KafkaUser
		c.Password = cache.GetPassword(ctx, p.KafkaPass, "kafkaPass")
	}

	switch p.KafkaAuthType {
	case model.KafkaAuthTypeSASLKerberos:
		name1 := getName(ctx.TraceVal, keyTab)
		// 先删除之前的文件
		files, err := os.ReadDir("./")
		if err == nil {
			for _, v := range files {
				if v.IsDir() {
					continue
				}
				_ = deleteFile(v.Name())
			}
		}

		if err := createFile(p.KafkaKeytab, name1); err != nil {
			return nil, errors.WithMessagef(err, "create keytab file failed")
		}

		defer func() {
			if err := deleteFile(name1); err != nil {
				ctx.Errorf(" delete file failed, fileName %s", name1)
			}
		}()

		name2 := getName(ctx.TraceVal, kerberosConf)
		if err := createFile(p.KafkaKrb5Conf, name2); err != nil {
			return nil, errors.WithMessagef(err, "create keytab file failed")
		}
		defer func() {
			if err := deleteFile(name2); err != nil {
				ctx.Errorf("delete file failed, fileName %s", name1)
			}
		}()

		c.SaslEnable = true
		c.ServiceName = p.KafkaKrbServiceName
		c.Realm = p.KafkaKrbRealm
		c.KeyTabPath = "./" + name1
		c.KerberosConfigPath = "./" + name2
		c.Password = p.KafkaPass
		c.Algorithm = sarama.SASLTypeGSSAPI

	case model.KafkaAuthTypeSASLGssapi:
		c.Algorithm = "gssapi"
		c.SaslEnable = true
	case model.KafkaAuthTypeSha512:
		c.Algorithm = "sha512"
		c.SaslEnable = true
	case model.KafkaAuthTypeSha256:
		c.Algorithm = "sha256"
		c.SaslEnable = true
	}
	//ctx.Infof("kafka param %#v", p)
	ctx.Debugf("kafka config %#v", *c)
	return c, nil
}

// deleteFile 要求只能删除keytab 及 conf 文件
func deleteFile(name string) error {
	switch name {
	case "./", "/":
		return nil
	}

	if strings.Contains(name, "..") {
		return nil
	}
	if strings.Contains(name, "*") {
		return nil
	}
	if !strings.Contains(name, "keytab") || !strings.Contains(name, "conf") {
		return nil
	}

	if info, err := os.Stat(name); err == nil {
		if info.IsDir() {
			return nil
		}
		return os.RemoveAll(name)
	} else {
		if os.IsNotExist(err) {
			return nil
		}
	}
	return nil
}

// getName 生成临时文件名称
func getName(requestId string, fileType string) string {
	return fmt.Sprintf("%s_%s_%s", fileType, requestId, time.Now().Format("2006-01-02T15-04-05.999999Z07-00"))
}

// createFile 因为配置需要传入文件地址，且没有找直接传入文件内容的方法
func createFile(data string, fileName string) error {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return errors.WithMessage(err, "base64 decode failed")
	}
	bStr := string(b)
	lst := strings.Split(bStr, "base64,")
	if len(lst) != 2 {
		return errors.Wrapf(err, "invalide base64 info %s", bStr)
	}
	b, err = base64.StdEncoding.DecodeString(lst[1])
	if err != nil {
		return errors.WithMessage(err, "base64 decode failed")
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

const (
	KeyRuleID   = "RuleId"
	KeyRuleType = "RuleType"
)

func (c *ReceiverKafka) Close() error {
	c.cancelFunc()
	return nil
}

func (c *ReceiverKafka) Collect(ruleId string, param map[string]interface{}, f model2.HandleFunc) (cf func(), err error) {
	c.ctx.Infof("kafka param %#v", param)
	conf, err := genKafkaConf(c.ctx, param)
	if err != nil {
		return nil, errors.Wrapf(err, "gen kafka conf")
	}

	if c.WorkSize == 0 {
		c.WorkSize = 100
	}
	//ctx_, cf_ := context.WithCancel(context.TODO())
	//c.cf = cf_
	//go func() {
	//var kafkaConsumer *kafka.Consumer
	//c.ctx.Infof("kafka config %#v", conf)
	c.ctx.Context = context.WithValue(c.ctx, KeyRuleID, ruleId)

	return client.NewKafkaConsumerCli(c.ctx, conf, f, client.KafkaArg{
		RecvSize: c.WorkSize,
		WorkSize: c.WorkSize,
	})
	//kafkaConsumer, err = kafka.NewConsumer(conf)
	//if err != nil {
	//	kctx.Errorf("Failed to create Kafka consumer,waiting s and retry... ... \n err: %v", err)
	//	time.Sleep(time.Second * time.Duration(c.RetryIntervalSecond))
	//	cf, err = c.Collect(ruleId, param, f)
	//	return
	//}

	//ctx := context.WithValue(context.TODO(), KeyRuleID, ruleId)
	//ctx = context.WithValue(ctx, KeyRuleType, model.RuleTypeKafka)
	//c2 := kctx.NewContext(ctx, util.NewUUIDStr())
	//c.consumer = kafkaConsumer
	//
	//c.cancelFunc, err = kafkaConsumer.Consume(c2, f)
	//select {
	//case <-ctx_.Done():
	//	c.ctx.Infof("goroutine exsit for kafka collect")
	//	return
	//}
	//}()

	return func() {
		c.Close()
	}, nil
}
