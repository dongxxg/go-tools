package model

import (
	"unitechs.com/cno/kit/lib/mq/model"
)

type IReceiver interface {
	Collect(ruleId string, param map[string]interface{}, f model.HandleFunc) (cf func(), err error)
	Close() error
}

type Rule struct {
	RuleId     string           `json:"ruleId"`
	RuleName   string           `json:"ruleName"`
	RuleType   RuleType         `json:"ruleType"`
	RuleDesc   string           `json:"ruleDesc"`
	RuleStatus int              `json:"ruleStatus"` // 0 停用  1 启用
	TemplId    string           `json:"templId"`    // 调用31395 的模板redis 缓存  该id值为 subType  ， tempType 固定为 subscribe
	FilterList FilterList       `json:"filterList"`
	SubParas   SubParas         `json:"subParas"`
	HandleFunc model.HandleFunc `json:"-"`
	Receiver   IReceiver        `json:"-"`
}

// RuleType 规则类型， 即发送的目的地
type RuleType string

func (t RuleType) String() string {
	return string(t)
}

type ValueType string

func (t ValueType) String() string {
	return string(t)
}

const (
	ValueTypeString    ValueType = "string"
	ValueTypeNumber              = "number"
	ValueTypeDate1               = "date(yyyy-mm-dd hh:mi:ss)"
	ValueTypeDate2               = "date(yyyymmddhhmiss)"
	ValueTypeDateStamp           = "date(timestamp)"
)

const (
	RuleTypeKafka          RuleType = "kafka"
	RuleTypeCTGMQ          RuleType = "ctg-mq"
	RuleTypeRest           RuleType = "rest"
	RuleTypeKafkaPrehandle RuleType = "kafka_prehandle"

	RuleTypeEmail  RuleType = "email"
	RuleTypeWeChat RuleType = "wechat"
	RuleTypeES     RuleType = "es"
	RuleTypeSMS    RuleType = "sms"
	RuleTypeSocket RuleType = "socket"
	RuleTypePrint  RuleType = "print" // RuleTypePrint 打印 测试用
)

type SubParas map[string]interface{}

type FilterList []Filter

func (l FilterList) IsMatched() bool {
	for _, v := range l {
		if v.IsMatched {
			return true
		}
	}

	return false
}

// DelayTime get delay time
// return Seconds to delay if needed. else return -1
func (l FilterList) DelayTime() int {
	t := -1
	for _, v := range l {
		if v.IsDelay == 1 {
			if v.DelayTime > 0 && v.DelayTime > t {
				t = v.DelayTime
			}
		}
	}
	return t
}

type KafkaAuthType string

// update cm_ala_subscribe_dict set dict_list='ASL/PLAIN:SASL/PLAIN,SASL/Kerberos:SASL/Kerberos,SASL/SHA512:SASL/SHA512,SASL/SHA256:SASL/SHA256,SASL/GSSAPI:SASL/GSSAPI'
// where sub_type='kafka' and para_code='kafkaAuthType'
const (
	KafkaAuthTypeSASLPlain    KafkaAuthType = "SASL/PLAIN"
	KafkaAuthTypeSASLKerberos KafkaAuthType = "SASL/Kerberos"
	KafkaAuthTypeSha512       KafkaAuthType = "SASL/SHA512"
	KafkaAuthTypeSha256       KafkaAuthType = "SASL/SHA256"
	KafkaAuthTypeSASLGssapi   KafkaAuthType = "SASL/GSSAPI"
)

// ParamKafka 需要兼容
type ParamKafka struct {
	KafkaBroker   string        `json:"RECV_KAFKA_BROKERS"`
	KafkaTopic    string        `json:"RECV_KAFKA_TOPIC"`
	KafkaUser     string        `json:"RECV_KAFKA_SASL_USERNAME"`
	KafkaPass     string        `json:"RECV_KAFKA_SASL_PASSWORD"`
	KafkaAuthType KafkaAuthType `json:"RECV_KAFKA_SASL_ALGORITHM"` // 认证机制
	KafkaKeytab   string        `json:"kafkaKeytab"`               // base64 加密的文件
	KafkaKrb5Conf string        `json:"kafkaKrb5conf"`             // base64 加密的文件
	KafkaHeader   string        `json:"kafkaHeader"`               // 消息头， json 字符串
	KafkaGroup    string        `json:"RECV_KAFKA_GROUP"`

	// 后面新加的， 原来没有
	KafkaKrbServiceName string `json:"kafkaKrbServiceName"` // 服务名称
	KafkaKrbRealm       string `json:"KafkaKrbRealm"`       // 域名
	KafkaVersion        string `json:"KafkaVersion"`        // 版本号
}
type ParamCTGMQ struct {
	CTGmqBroker string `json:"ctgmqBroker"` // 服务地址
	CTGmqTopic  string `json:"ctgmqTopic"`  // 订阅主题
	CTGmqUser   string `json:"ctgmqUser"`   //
	CTGmqPass   string `json:"ctgmqPass"`   //
	CTGmqGroup  string `json:"ctgmqgroup"`
}

type ParamRest struct {
	PubRestUrl           string `json:"PUB_REST_URL"`
	PubRestMethod        string `json:"PUB_REST_METHOD"`
	PubRestHeader        string `json:"PUB_REST_HEADER"` // X
	PubRestDataField     string `json:"PUB_REST_DATA_FIELD"`
	PubRestPeriod        string `json:"PUB_REST_PERIOD"`
	PubRestPage          string `json:"PUB_REST_PAGE"`
	PubRestStatusField   string `json:"PUB_REST_STATUS_FIELD"`
	PubRestStatusValue   string `json:"PUB_REST_STATUS_VALUE"`
	PubRestPagesizeField string `json:"PUB_REST_PAGESIZE_FIELD"`
	PubRestCurPage       string `json:"PUB_REST_CUR_PAGE"`
	PubRestAppid         string `json:"PUB_REST_APPID"`
	PubRestAppSecret     string `json:"PUB_REST_APPSECRET"`
	PubRestParam         string `json:"PUB_REST_PARA"`
}

type ParamKafkaPrehandle struct {
	RecvKafkaPrehandleBrokers       string `json:"RECV_KAFKA_PREHANDLE_BROKERS"`
	RecvKafkaPrehandleTopic         string `json:"RECV_KAFKA_PREHANDLE_TOPIC"`
	RecvKafkaPrehandleGroup         string `json:"RECV_KAFKA_PREHANDLE_GROUP"`
	RecvKafkaPrehandleSaslUsername  string `json:"RECV_KAFKA_PREHANDLE_SASL_USERNAME"`
	RecvKafkaPrehandleSaslPassword  string `json:"RECV_KAFKA_PREHANDLE_SASL_PASSWORD"`
	RecvKafkaPrehandleSaslEnabled   string `json:"RECV_KAFKA_PREHANDLE_SASL_ENABLED"`
	RecvKafkaPrehandleSaslAlgorithm string `json:"RECV_KAFKA_PREHANDLE_SASL_ALGORITHM"`
	// type=AlarmStatusChange.AlarmRelation
	RecvKafkaPrehandleField string `json:"RECV_KAFKA_PREHANDLE_FIELD"` // type
	RecvKafkaPrehandleValue string `json:"RECV_KAFKA_PREHANDLE_VALUE"` // AlarmStatusChange.AlarmRelation
	RecvAlarmService        string `json:"RECV_ALARM_SERVICE"`         // 告警服务名称
	RecvMsgField            string `json:"RECV_MSG_FIELD"`             // msg
	RecvMsgPath             string `json:"RECV_MSG_PATH"`              // alarmUniqueClearId
}
