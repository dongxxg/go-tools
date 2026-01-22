// 2022-03-29 文康  32032_230326_中盈统一业务应用中台-32032-告警接收模块支持规则页面可配置-闵莽锐

package rule

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"unitechs.com/cno/kit/kctx"

	model2 "unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/template/cache"
	"unitechs.com/cno/kit/template/model"
	receiver2 "unitechs.com/cno/kit/template/rule/receiver"
	"unitechs.com/cno/kit/util"
)

var receiver *Receiver

func GetReceiver() *Receiver {
	return receiver
}

func GetRuleById(id string) (model.Rule, bool) {
	return receiver.GetRuleById(id)

}
func GetTemplateById(ruleId string) (string, error) {
	r, ok := receiver.rules.Get(ruleId)
	if !ok {
		return "", fmt.Errorf("ruleId not exist %s", ruleId)
	}

	return r.TemplId, nil
}
func GetTemplate(ruleId, prefix string) (model.ConvertRule, error) {
	tmpId, err := GetTemplateById(ruleId)
	if err != nil {
		return model.ConvertRule{}, errors.Wrapf(err, "failed to GetTemplateById, ruleId %s", tmpId)
	}
	if !strings.HasPrefix(tmpId, prefix) {
		tmpId = prefix + tmpId
	}

	t, ok := receiver.templates.Get(tmpId)
	if !ok {
		return model.ConvertRule{}, fmt.Errorf("get template by id failed, templateId %s", tmpId)
	}
	return t, nil
}

// Receiver 接收器
type Receiver struct {
	rules     *cache.RuleCache       // 规则
	templates *cache.ConvertTemplate // 模板
	ctx       kctx.Ctx
	arg       ReceiverArg
}

type ReceiverArg struct {
	RestArg     RestArg
	RestServArg RestServArg
	PoolSize    int // kafka 工作线程数
}

type RestArg struct {
	BreakRepo receiver2.IBreakTime
	PoolSize  int // rest 处理任务协程池的大小
}
type RestServArg struct {
	Uri      string //  服务在地址
	PoolSize int    // rest 处理任务协程池的大小
}

func NewReceiver(ctx kctx.Ctx, rules *cache.RuleCache, templates *cache.ConvertTemplate, arg ReceiverArg) *Receiver {
	r := &Receiver{
		rules:     rules,
		templates: templates,
		ctx:       ctx,
		arg:       arg,
	}
	receiver = r
	return r
}
func (c *Receiver) RunBy(ruleId string) (cf func(), err error) {
	r, ok := c.GetRuleById(ruleId)
	if ok && r.Receiver != nil {
		go func() {
			cf, err = r.Receiver.Collect(r.RuleId, r.SubParas, r.HandleFunc)
			if err != nil {
				c.ctx.Errorf("runnning server failed %s ruleId %s, err: %v", r.RuleType, r.RuleId, err)
			} else {
				c.ctx.Infof("running server success %s ruleId %s", r.RuleType, r.RuleId)
			}
		}()
		return
	}
	err = fmt.Errorf("handler/receiver  of rule not exist")
	return
}

// RegisterByType 这里是几种常见的方式， 内置在接收器中了， 可以注册其他的接收方式
func (r *Receiver) RegisterByType(handleMap map[model.RuleType]model2.HandleFunc) {
	r.rules.Range(func(rule2 *model.Rule) bool {
		switch rule2.RuleType {

		case model.RuleTypeCTGMQ:
			h, ok := handleMap[model.RuleTypeCTGMQ]
			if !ok {
				r.ctx.Errorf("rule type %s not exist in rule, no need to registry", rule2.RuleType)
				return true
			}
			//p := model.ParamCTGMQ{}
			//if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
			//	r.ctx.Errorf("failed to parse %s param, err: %v, param: %+v", rule2.RuleType, err, rule2.SubParas)
			//	return true
			//}
			//brokers := util.SplitBy(p.CTGmqBroker, ",")
			//p.CTGmqPass = cache.GetPassword(r.ctx, p.CTGmqPass, "ctgmqPass")
			//r.ctx.Debugf("%s param %+v", rule2.RuleType, p)
			recv, err := receiver2.NewReceiverCTGMQWith(r.ctx, 0)
			if err != nil {
				r.ctx.Errorf("failed to NewCTGMQReceiverWith, err: %v", err)
			}

			r.RegisterHandleByID(rule2.RuleId, h, recv)

		case model.RuleTypeKafka:
			h, ok := handleMap[model.RuleTypeKafka]
			if !ok {
				r.ctx.Errorf("rule type %s not exist in rule, no need to registry", rule2.RuleType)
				return true
			}
			//p := model.ParamKafka{}
			//if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
			//	r.ctx.Errorf("failed to parse %s param, err: %v, param: %+v", rule2.RuleType, err, rule2.SubParas)
			//	return true
			//}
			//p.KafkaPass = cache.GetPassword(r.ctx, p.KafkaPass, "RECV_KAFKA_SASL_PASSWORD")
			//r.ctx.Debugf("%s param %+v", rule2.RuleType, p)
			recv, err := receiver2.NewReceiverKafka(r.ctx, r.arg.PoolSize)
			if err != nil {
				r.ctx.Errorf("failed to NewReceiverKafkaWith, err: %v", err)
			}

			r.RegisterHandleByID(rule2.RuleId, h, recv)

		case model.RuleTypeKafkaPrehandle:
			h, ok := handleMap[model.RuleTypeKafkaPrehandle]
			if !ok {
				r.ctx.Errorf("rule type %s not exist in rule, no need to registry", rule2.RuleType)
				return true
			}
			p := model.ParamKafkaPrehandle{}
			if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
				r.ctx.Errorf("failed to parse %s param, err: %v, param: %+v", rule2.RuleType, err, rule2.SubParas)
				return true
			}
			r.ctx.Debugf("%s param %+v", rule2.RuleType, p)
			if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
				r.ctx.Errorf("failed to parse prehandle param, err: %v, param: %+v", err, rule2.SubParas)
				return true
			}
			p.RecvKafkaPrehandleSaslPassword = cache.GetPassword(r.ctx, p.RecvKafkaPrehandleSaslPassword, "RECV_KAFKA_PREHANDLE_SASL_PASSWORD")
			recv, err := receiver2.NewReceiverPrehandle(r.ctx, p)
			if err != nil {
				r.ctx.Errorf("failed to NewReceiverPrehandle, err: %v", err)
			}
			r.RegisterHandleByID(rule2.RuleId, h, recv)

		case model.RuleTypeRest:
			h, ok := handleMap[model.RuleTypeRest]
			if !ok {
				r.ctx.Errorf("rule type %s not exist in rule, no need to registry", rule2.RuleType)
				return true
			}
			p := model.ParamRest{}
			if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
				r.ctx.Errorf("failed to parse %s param, err: %v, param: %+v", rule2.RuleType, err, rule2.SubParas)
				return true
			}
			r.ctx.Debugf("%s param %+v", rule2.RuleType, p)
			if err := util.MarshalWith(rule2.SubParas, &p); err != nil {
				r.ctx.Errorf("failed to parse %s param, err: %v, param: %+v", rule2.RuleType, err, rule2.SubParas)
				return true
			}
			rest, err := receiver2.NewReceiverRestWith(r.ctx, p, r.arg.RestArg.PoolSize, r.arg.RestArg.BreakRepo)
			if err != nil {
				r.ctx.Errorf("failed to NewReceiverRestWith, err: %v", err)
			}
			r.RegisterHandleByID(rule2.RuleId, h, rest)
			//default:
			//	r.ctx.Errorf("not supported rule type %s", rule2.RuleType)
		}

		return true
	})
	fmt.Println("here")
}

func (r *Receiver) Stop(ctx kctx.Ctx) {
	r.rules.Range(func(rule2 *model.Rule) bool {
		if rule2.Receiver != nil {
			if err := rule2.Receiver.Close(); err != nil {
				ctx.Errorf("rule %s  ruleType: %s close failed: %v", rule2.RuleId, rule2.RuleType, err)
			} else {
				ctx.Infof("rule %s  ruleType: %s close success", rule2.RuleId, rule2.RuleType)
			}
		}
		return true
	})
	ctx.Infof("all rule closed")
}
func (r *Receiver) Run() func() {
	//cfList := make([]func(), 0)
	//cfList := gset.NewSet(true)
	cfList := sync.Map{}
	r.rules.Range(func(rule2 *model.Rule) bool {
		if rule2 == nil {
			return true
		}

		go func(r_ *model.Rule) {
			var err error
			var cf func()
			defer util.HandlePanic(r.ctx, "error", func() {
				r.ctx.Errorf("runnning server failed %s ruleId %s", r_.RuleType, r_.RuleId)
			})
			if r_.Receiver == nil {
				r.ctx.Errorf("receiver is nil for ruleType %s ruleId %s", r_.RuleType, r_.RuleId)
				return
			}

			cf, err = r_.Receiver.Collect(r_.RuleId, r_.SubParas, r_.HandleFunc)
			if err != nil {
				r.ctx.Errorf("runnning server failed %s ruleId %s, err: %v", r_.RuleType, r_.RuleId, err)
			} else {
				r.ctx.Infof("running server success %s ruleId %s", r_.RuleType, r_.RuleId)
			}
			cfList.Store(util.NewUUIDStr(), cf)

		}(rule2)

		return true
	})

	return func() {
		cfList.Range(func(key, value interface{}) bool {
			cf, ok := value.(func())
			if ok {
				cf()
			}
			return true
		})
	}

}
func (r *Receiver) RegisterHandleByType(ruleType model.RuleType, f model2.HandleFunc, receiver model.IReceiver) {
	r.rules.Range(func(rule2 *model.Rule) bool {

		if rule2.RuleType == ruleType {
			rule2.HandleFunc = f
			rule2.Receiver = receiver
			//r.rules.Set(rule2.RuleId, rule2)
		}
		return true
	})
}

func (r *Receiver) RegisterHandleByID(ruleID string, f model2.HandleFunc, receiver model.IReceiver) {
	r.rules.Range(func(rule2 *model.Rule) bool {
		if rule2.RuleId == ruleID {
			rule2.HandleFunc = f
			rule2.Receiver = receiver
			return false
		}
		return true
	})
}

// CustomRegister 自定义注册
func (r *Receiver) CustomRegister(f func(rule *model.Rule) bool) {
	r.rules.Range(f)
}

func (r *Receiver) GetRuleById(id string) (model.Rule, bool) {
	return r.rules.Get(id)
}
func (r *Receiver) DeleteRule(ruleId string) {
	rule, ok := r.rules.Get(ruleId)
	if ok && rule.Receiver != nil {
		if err := rule.Receiver.Close(); err != nil {
			r.ctx.Errorf("runnning server failed %s ruleId %s, err: %v", rule.RuleType, rule.RuleId, err)
		} else {
			r.ctx.Infof("running server success %s ruleId %s", rule.RuleType, rule.RuleId)
		}
	}
	r.rules.Delete(ruleId)
}
func (r *Receiver) DeleteTemplate(templId string) {
	r.templates.Delete(templId)
}

func (r *Receiver) SetRule(ctx kctx.Ctx, ruleMap map[string]string) (ruleIds []string) {
	return r.rules.SetByMap(ctx, ruleMap)
}

func (r *Receiver) SetTemplate(ctx kctx.Ctx, tmplMap map[string]string) {
	r.templates.SetByMap(ctx, tmplMap)
}
