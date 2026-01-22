// 任务号    日期     开发者      描述

package alarm

import (
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tidwall/sjson"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/db/kredis"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
)

const (
	NotFoundPolicyEmpty = "empty" // 当找不到时， 设置为空
	NotFoundPolicyExpr  = "expr"  // 当找不到时， 显示为公式表达式， 如${k_transcode}
)

// TemplateConvertorOption 维护一个模板转换池
// 模板转换工具
type TemplateConvertorOption struct {
	NotFoundPolicy string // 当消息中没有找到模板配置的值的时候的策略
	ApiCache       ApiCacheOption
	SerialNo       SerialNoOption
}
type SerialNoOption struct {
	RedisName string
	Start     int64
	Step      int64
}

type ResType string

const (
	ResTypeDev  ResType = "DEV"
	ResTypePort ResType = "PORT"
	ResTypeCIR  ResType = "CIR"
	ResTypePATH ResType = "PATH"
	ResTypeCard ResType = "Card"
)

var tempConvertor *TemplateConvertor

func SetTempConvertor(c *TemplateConvertor) {
	tempConvertor = c
}

// MatchAndConvert  匹配和转换
/*
alarm.MatchAndConvert(ctx, msg, ConvertTmpl, alarm.MatchArg{
		RuleId:     ruleID,
		TemplateId: "empty_tmpl_id",
		FindAttrFunc: alarm.FindResAttrFunc(map[string]string{}, map[string]string{}),
	})


*/
func MatchAndConvert(ctx kctx.Ctx, m model.IMsg, r model.ConvertRule, arg MatchArg) (string, error) {
	if tempConvertor == nil {
		return "", fmt.Errorf("tempate convertor is nil, ")
	}
	return tempConvertor.MatchAndConvert(ctx, m, r, arg), nil
}

// TemplateConvertor 模板转换器
type TemplateConvertor struct {
	Option     TemplateConvertorOption
	cache      model.ICache // 缓存器
	SerialNo   ISerialNo
	ExpireHour int    //apiCache 缓存到期时间
	Zone       string // 时间 默认Local
	writeLock  sync.RWMutex
}

func (t *TemplateConvertor) Copy() *TemplateConvertor {
	return &TemplateConvertor{
		Option:     t.Option,
		cache:      t.cache,
		SerialNo:   t.SerialNo,
		ExpireHour: t.ExpireHour,
		Zone:       t.Zone,
		writeLock:  sync.RWMutex{},
	}
}

func (t *TemplateConvertor) SetZone(zone string) {
	t.Zone = zone
}

func NewTemplateConvertor(option TemplateConvertorOption, cache model.ICache) *TemplateConvertor {
	return &TemplateConvertor{Option: option, cache: cache, Zone: "Local"}
}

func NewTemplateConvertorWith(ctx kctx.Ctx, repo *kredis.RedisRepo, option TemplateConvertorOption) TemplateConvertor {
	//serialNo := kredis.NewSerialNoRepo(repo)
	serialNo := NewNumberGenerator(
		option.SerialNo.Start,
		option.SerialNo.Step,
		repo.GetDB(ctx),
		option.SerialNo.RedisName)

	return TemplateConvertor{
		Option:   option,
		cache:    repo,
		SerialNo: serialNo,
	}
}

func NewTemplateConvertorWithRedis(ctx kctx.Ctx, Type, host, pwd string, db int, option TemplateConvertorOption) (templ *TemplateConvertor, cf func(), er error) {
	repo, err := kredis.NewRedisRepo(ctx, Type, host, pwd, db)
	if err != nil {
		return nil, func() {}, errors.Wrapf(err, "new redis repo failed err: %v", err)
	}

	//serialNo := kredis.NewSerialNoRepo(repo)

	serialNo := NewNumberGenerator(
		option.SerialNo.Start,
		option.SerialNo.Step,
		repo.GetDB(ctx),
		option.SerialNo.RedisName)

	if option.ApiCache.ExpireHour == 0 {
		option.ApiCache.ExpireHour = 24
	}
	r := TemplateConvertor{
		Option:     option,
		cache:      repo,
		SerialNo:   serialNo,
		ExpireHour: option.ApiCache.ExpireHour,
	}

	return &r, cf, nil
}

// GetSerialNo 获取序列号
// ruleid, template, dstField 都是查询全局锁需要的字段
// GenSerialNoLockKey 生成序列号在redis中存储的锁的key
//
//	func GenSerialNoLockKey(ruleId, templateId, dstField string) string {
//		return fmt.Sprintf("serial:%s:%s:%s:lock", ruleId, templateId, dstField)
//	}
func (t *TemplateConvertor) GetSerialNo(ctx kctx.Ctx, ruleId, templateId, dstField string, start, interval int64) (int64, error) {
	return t.SerialNo.Get(ctx, ruleId, templateId, dstField, start, interval)
}

func (t *TemplateConvertor) MatchWithMapOut(ctx kctx.Ctx, m model.IMsg, r model.ConvertRule, arg MatchArg) (map[string]string, error) {
	s := t.MatchAndConvert(ctx, m, r, arg)
	if s == "" {
		return nil, nil
	}
	out := make(map[string]string)

	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, errors.Wrapf(err, "unmarshal to map failed,  %s", s)
	}

	return out, nil
}

// MatchAndConvert 消息匹配规则并按模板转换
// 如果有多个， 则返回json 字符串
func (t *TemplateConvertor) MatchAndConvert(ctx kctx.Ctx, m model.IMsg, r model.ConvertRule, arg MatchArg) string {
	var out string

	hasMatched := sync.Map{}

	tt := t.Copy()

	for i, v := range r.Map {
		// 同一个字段, 匹配到一个, 则停止
		if _, ok := hasMatched.Load(v.DstField); ok {
			continue
		}
		// 该字段是否符合条件

		condKey, _, ok := m.GetCondKey()
		if !ok {
			ctx.Errorf("cond key %s not exist in msg", condKey)
			continue
		}
		condVal, err := t.GetAttr(condKey, m, out)
		if err != nil {
			ctx.Errorf("get cond val failed: %v", err)
		}
		condValStr := cast.ToString(condVal)

		typeIds := gset.NewStrSetFrom(strings.Split(v.AlarmType.String(), ","))
		typeIds.Remove("")

		if typeIds.Contains(condValStr) || v.AlarmType == model.TypeAll {
			ctx.Infof("alarm type matched index %d,dst_field '%s', srcField '%s', expr '%s', alarmType '%s'", i, v.DstField, v.SrcField, v.MapExpr, v.AlarmType)
			if v.MapType == model.MapTypeDict {
				if err := tt.MatchWithType(ctx, m, v, r.Dict, &out, arg); err != nil {
					ctx.Errorf("MatchWithType failed, index %d,  rule %v,  err: %v", i, v, err)
				}
			} else {
				if err := tt.MatchWithType(ctx, m, v, model.Dicts{}, &out, arg); err != nil {
					ctx.Errorf("MatchWithType failed, index %d,  rule %v,  err: %v", i, v, err)
				}
			}
			//ctx.Infof("single match time %d index %d, %v", time.Now().Sub(n).Milliseconds(), i, v)
		} else {
			//ctx.Errorf("alarm type not matched index %d, rule %s,  msg %s ", i, v.AlarmType, condVal)
			//ctx.Infof("single match time %d index %d, %v", time.Now().Sub(n).Milliseconds(), i, v)
		}
	}
	//wg.Wait()
	return out
}

type MatchArg struct {
	RuleId       string    // serialNo 的时候需要传入
	TemplateId   string    // template_id 模板id
	FindAttrFunc FindAttrF // 找属性的函数  查找findAttr 的具体实现 已经内置方法可选择 FindResAttrFunc
}

type FindAttrFun func(devType string, cond map[string]interface{}, dstField string) (interface{}, error)

func (t *TemplateConvertor) MatchWithType(ctx kctx.Ctx, m model.IMsg, v model.Elem, d model.Dicts, s *string, arg MatchArg) error {
	var (
		val interface{}
		err error
	)
	switch v.MapType {
	case model.MapTypeFixed:
		val, err = t.FixedFunc(ctx, v, m)
	case model.MapTypeDirect:
		val, err = t.DirectFunc(ctx, v, m, s)
	case model.MapTypeJoin:
		val, err = t.JoinFunc(ctx, v, m, *s)
	case model.MapTypeAPI:
		val, err = t.APIFunc(ctx, v, m, *s)
	case model.MapTypeCase:
		val, err = t.CaseFunc(ctx, v, m, *s)
	case model.MapTypeDict:
		val, err = t.DictFunc(ctx, v, m, d, *s)
	case model.MapTypeNvl:
		val, err = t.NvlFunc(ctx, v, m, *s)
	case model.MapTypeSerial:
		val, err = t.SerialFunc(ctx, v, m, arg.RuleId, arg.TemplateId, *s, t.SerialNo)
	case model.MapTypeReg:
		val, err = t.RegFunc(ctx, v, m, *s)
	case model.MapTypeRegCase:
		val, err = t.RegCaseFunc(ctx, v, m, *s)
	case model.MapTypeFindAttr:
		var resClass string
		resClass, val, err = t.FindAttrFunc(ctx, v, arg.FindAttrFunc)
		if err != nil && strings.Contains(err.Error(), "not exist") {
			// l_开头的PORT 赋值， 如果不存在， 则赋值0
			switch ResType(resClass) {
			case ResTypePort, ResTypeCIR, ResTypePATH, ResTypeCard:
				if strings.HasPrefix(v.DstField, "l_") && val == nil {
					val = "0"
					err = nil
				}
			}

		}
	case model.MapTypeApiCache:
		val, err = t.ApiCache(ctx, v, m, *s)
	case model.MapTypeTimeStdFormat:
		val, err = t.DirectFunc(ctx, v, m, s)
		val = t.parseTime(ctx, val, util.TimeFormatDefault)
	case model.MapTypeTimeRfc3339:
		val, err = t.DirectFunc(ctx, v, m, s)
		val = t.parseTime(ctx, val, time.RFC3339Nano)
	case model.MapTypeTimeCompactFormat:
		val, err = t.DirectFunc(ctx, v, m, s)
		val = t.parseTime(ctx, val, util.TimeFormatSec)
	case model.MapTypeTimestamp:
		val, err = t.DirectFunc(ctx, v, m, s)
		val = t.parseTime(ctx, val, util.TimeFormatTimestamp)
	case model.MapTypeAutoSwitch:
		val, err = t.AutoSwitchFunc(ctx, v, m, s)

	default:
		return fmt.Errorf("invalid mapType %s", v.MapType)
	}
	if err != nil {
		ctx.Errorf("convert failed, type %s, %v", v.MapType, err)
		val = ""
	}
	t.writeLock.Lock()
	s_, err := sjson.Set(*s, v.DstField, val)
	t.writeLock.Unlock()
	ctx.Infof("set '%s'='%s'", v.DstField, val)
	if err != nil {
		return fmt.Errorf("set json failed, err: %s", err.Error())
	}
	*s = s_
	return nil
}
func (t *TemplateConvertor) parseTime(ctx kctx.Ctx, val interface{}, timeFormat string) interface{} {
	if val != nil {
		if valS, ok := val.(string); ok {
			loc, err := time.LoadLocation(t.Zone)
			if err != nil {
				ctx.Errorf("invalid zone str %s, err: %v", t.Zone, err)
			} else {
				tt, err := util.ParseToTimeWithZone(valS, loc)
				if err != nil {
					ctx.Errorf("invalid time str, %s", valS)
				} else {
					switch timeFormat {
					case util.TimeFormatTimestamp:
						return tt.Unix()
					default:
						val = tt.Format(timeFormat)
					}
				}
			}
		}
	}
	return val
}
