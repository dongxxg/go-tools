// 任务号    日期     开发者      描述

package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"hash/crc32"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/cache"
	"unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
	"unitechs.com/unios-dice/uni-base/core/config"
)

// checkTemplateDetail 检测模板参数
func checkTemplateDetail(tmpl *model.Elem) error {
	if tmpl.DstField == "" {
		return fmt.Errorf("DstField can not be empty, desc=%s, ", tmpl.SrcField)
	}
	return nil
}

// FixedFunc fixed-固定值：直接赋值为一个常量
func (t *TemplateConvertor) FixedFunc(ctx kctx.Ctx, e model.Elem, m model.IMsg) (val interface{}, err error) {
	if err = checkTemplateDetail(&e); err != nil {
		return nil, err
	}
	return e.MapExpr, nil
}

func (t *TemplateConvertor) RegFunc(ctx kctx.Ctx, e model.Elem, m model.IMsg, s string) (val interface{}, err error) {
	if err = checkTemplateDetail(&e); err != nil {
		return nil, err
	}
	//key = e.DstField

	var v_ interface{}

	v_, err = t.GetAttr(e.SrcField, m, s)
	if err != nil {
		return nil, errors.Wrapf(err, "RegFunc get val failed, not exist key %s", e.SrcField)
	}

	strStr := util.ConvertToString(v_)

	r, err := regexp.Compile(e.MapExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "reg compile failed, err: %v", err)
	}
	lst := r.FindAllStringSubmatch(strStr, -1)
	// 不管什么情况， 只获取匹配到的第一个值
	for _, v := range lst {
		if v == nil {
			continue
		}
		for k, v2 := range v {
			if k == 0 {
				continue
			}
			if v2 == "" {
				continue
			}
			val = v2
			return
		}
		continue
	}
	// 没有匹配到内容
	err = fmt.Errorf("nothing mathced, srcInfo: %s, regexp: %s", e.SrcField, e.MapExpr)
	return
}

// caseRes expr=1:a
// attr expr  val = 1  result = a
type caseRes struct {
	attr   string // 属性
	val    string // 表达式等于的值
	result string // 最后应该赋予的值
}

type caseResLst []caseRes

// IsKeyUnique case 模仿的是switch case 的用法， 只能有一个key值
func (c caseResLst) IsKeyUnique() (unique bool, keys string) {
	k := make(map[string]struct{})
	for _, v := range c {
		k[v.attr] = struct{}{}
	}
	delete(k, "default")

	if len(k) != 1 {
		s := ""
		for i, _ := range k {
			s += i + ","
		}
		s = strings.Trim(s, ",")

		return false, s
	}
	return true, ""
}
func (c caseResLst) GetDefault() *caseRes {
	for _, v := range c {
		if v.attr == "default" {
			return &v
		}
	}

	return nil
}

// parseCase 解析case 表达式
// 形如：expr=1:a;expr=2:b;c, 其中expr 代码属性值
// 兼容 ${name}=wiken:a;${name}=:b;c
// 兼容 ${RACK}=NA:1;${RACK}
func parseCase(expr string) caseResLst {
	// map[variable]caseRes
	res := make([]caseRes, 0)
	conds := strings.Split(expr, ";")
	for _, v := range conds {
		if !strings.Contains(v, ":") {
			// 没有： 说明是最后一个
			v = strings.Trim(v, " ")
			res = append(res, caseRes{
				attr:   "default",
				val:    "",
				result: v,
			})
			continue
		}
		if !strings.Contains(v, ":") && !strings.Contains(v, "=") {
			// : = 不同时包含->不合法->丢弃
			continue
		}
		l1 := strings.Split(v, ":")
		if len(l1) != 2 {
			continue
		}
		l2 := strings.Split(l1[0], "=")
		if len(l2) != 2 {
			continue
		}
		var key string
		varLst := getVariables(l2[0])
		if len(varLst) == 0 {
			key = l2[0]
		} else {
			key = varLst[0]
		}
		key = strings.Trim(key, " ")
		val := strings.Trim(l2[1], " ")
		val = strings.Trim(val, "\"")
		val = strings.Trim(val, "'")
		result := strings.Trim(l1[1], " ")
		result = strings.Trim(l1[1], "'")
		res = append(res, caseRes{
			attr:   key,
			val:    val,
			result: result,
		})
	}
	return res

}

// getVariables 获取变量名称
// vk "${name}_love ${name}"
// return []string{"name"}
func getVariables(s string) []string {
	r, _ := regexp.Compile(`\$\{(.*?)\}`)
	l := r.FindAllStringSubmatch(s, -1)

	if len(l) >= 1 {
		set := gset.NewStrSet()
		for _, v := range l {
			if len(v) > 1 {
				ss := v[1]
				ss = strings.TrimLeft(ss, "${")
				ss = strings.TrimRight(ss, "]")
				set.Add(ss)
			}
		}
		return set.Slice()
	}

	return nil
}

// CaseFunc 解析case 类型模板
func (t *TemplateConvertor) CaseFunc(ctx kctx.Ctx, temp model.Elem, m model.IMsg, s string) (val interface{}, err error) {
	if err = checkTemplateDetail(&temp); err != nil {
		return nil, err
	}
	conds := parseCase(temp.MapExpr)
	if len(conds) == 0 {
		err = fmt.Errorf("invalid case expr: %s", temp.MapExpr)
		return
	}
	isUnique, keys := conds.IsKeyUnique()
	if !isUnique {
		return "", fmt.Errorf("invalid case key: %s", keys)
	}
	for _, v := range conds {
		if v.attr == "default" {
			continue
		}
		var v_ interface{}

		v_, err := t.GetAttr(v.attr, m, s)
		if err != nil {
			ctx.Errorf("caseFunc %v", err)
			if t.Option.NotFoundPolicy == NotFoundPolicyExpr {
				continue
			}
		}
		if util.ConvertToString(v_) == v.val {
			if strings.HasPrefix(v.result, "${") {
				vs := getVariables(v.result)
				if len(vs) > 0 {
					v_, err = t.GetAttr(vs[0], m, s)
					if err == nil {
						v.result = cast.ToString(v_)
					}
				}
			}

			return v.result, nil
		}
	}

	d := conds.GetDefault()
	if d != nil {
		if strings.HasPrefix(d.result, "${") && strings.HasSuffix(d.result, "}") {
			return t.GetAttr(d.result[2:len(d.result)-1], m, s)
		}
		return d.result, nil
	}
	// 最后都没有匹配上， 返回空
	return nil, nil
}

func (t *TemplateConvertor) JoinFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s string) (interface{}, error) {
	if err := checkTemplateDetail(&temp); err != nil {
		return nil, err
	}
	res := temp.MapExpr

	varLst := getVariables(temp.MapExpr)

	for _, v := range varLst {
		var v_ interface{}
		v_, err := t.GetAttr(v, msg, s)
		if err != nil {
			ctx.Errorf("get val in result failed, key %s, err:%s", v, err)
			if t.Option.NotFoundPolicy == NotFoundPolicyExpr {
				continue
			}
		}
		s_ := util.ConvertToString(v_)
		res = strings.ReplaceAll(res, fmt.Sprintf("${%s}", v), s_)
	}
	return res, nil
}

func getValInJson(s, key string) (string, error) {
	r := gjson.Get(s, key)
	if !r.Exists() {
		return "", fmt.Errorf("key [%s] not exist in result", key)
	}
	return r.String(), nil
}

// setStrings 替换表达式中的变量
func (t *TemplateConvertor) setStrings(ctx kctx.Ctx, expr string, msg model.IMsg, s string) string {
	res := expr
	varLst := getVariables(expr)
	for _, v := range varLst {
		var v_ interface{}
		var err error
		v_, err = t.GetAttr(v, msg, s)
		if err != nil {
			ctx.Error(err)
			v_ = ""
		}
		s_ := util.ConvertToString(v_)
		res = strings.ReplaceAll(res, fmt.Sprintf("${%s}", v), s_)
	}

	return res
}

// NvlFunc 格式为： expr,a,b
// nvl-非空判断：此种情况下，映射表达式为条件表达式，格式为“expr,a,b”，即“表达式,不为空设值,为空设值”。
// 表示expr不为空则映射a，为空则映射b。a、b可以是变量也可以是常量。
func (t *TemplateConvertor) NvlFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s string) (interface{}, error) {
	l := strings.Split(temp.MapExpr, ",")
	if len(l) != 3 {
		return nil, fmt.Errorf("invalid Nvl mapExpr, format must be 'expr,a,b'. mapType=%s", temp.MapExpr)
	}
	fieldName := strings.Trim(l[0], " ")
	var v1, v2 string
	v1 = strings.Trim(l[1], " ")
	v1 = t.setStrings(ctx, v1, msg, s)

	v2 = strings.Trim(l[2], " ")
	v2 = t.setStrings(ctx, v2, msg, s)
	v_, err := t.GetAttr(fieldName, msg, s)
	if err != nil {
		ctx.Errorf("nvlFunc get field false, return default")
		// 字段没有也认为是空
		return v2, nil
	}

	if util.ConvertToString(v_) == "" {
		return v2, nil
	} else {
		return v1, nil
	}
}

// ISerialNo 获取序列号
type ISerialNo interface {
	Get(ctx kctx.Ctx, ruleId, templateId, dstField string, start, interval int64) (int64, error)
}

var serialNoPool = sync.Map{}

// SerialFunc 从redis 获取自增的序列号， 需要分布式锁
// 格式为“1,1”，即以1为起始值，每次加1
// serial-流水号：此种情况下，映射表达式为流水号的起始值、步长，逗号分隔。格式为“1,1”，即以1为起始值，每次加1。
func (t *TemplateConvertor) SerialFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, ruleId, templateId, s string, no ISerialNo) (interface{}, error) {
	if ruleId == "" || templateId == "" {
		return nil, fmt.Errorf("SerialFunc ruleId and templateId can`t be empty")
	}
	l := strings.Split(temp.MapExpr, ",")
	if len(l) != 2 {
		return nil, fmt.Errorf("invalid MapExpr for  serialFunc, MapExpr=%s, must such as  '1,1' ", temp.MapExpr)
	}
	start, err := strconv.ParseInt(l[0], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid MapExpr before ',' for SerialNo must be int %s", l)
	}
	interval, err := strconv.ParseInt(l[1], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid MapExpr before ',' for SerialNo %s", l)
	}
	res, err := no.Get(ctx, ruleId, templateId, temp.DstField, start, interval)
	if err != nil {
		return nil, errors.Wrapf(err, "SerialFunc get serialNo failed")
	}
	ss := util.ConvertToString(res)

	return ss, nil
}

// DirectFunc 直接获取变量
// direct-直接赋值：赋值为某个告警属性
func (t *TemplateConvertor) DirectFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s *string) (interface{}, error) {
	if err := checkTemplateDetail(&temp); err != nil {
		return nil, err
	}
	v_, err := t.GetAttr(temp.MapExpr, msg, *s)
	if err != nil {
		return nil, fmt.Errorf("get val in msg and result failed, err:%s", err)
	}

	return v_, nil
}

func (t *TemplateConvertor) APIFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s string) (interface{}, error) {
	_, val, err := t.apiFunc(ctx, temp, msg, s)
	return val, err
}

// apiFunc
/*
34856
api-API调用：调用指定API，入参为指定告警属性或变量，把调用接口的返回值进行映射。其中映射表达式为URL，采用get方法调用。
例如：http://RESOURCE-JAVA-SERVICE/api/ala/getCircuitNo?resid=${k_RESID}::outPara。其中“${}”为变量引用，
RESOURCE-JAVA-SERVICE为服务名；outPara为返参参数名，取值时取data下对应该参数名的参数值，若未指定，则直接取data的值。
返参格式如下，其中code=0000为成功、=0001为失败，data为返回值。
*/
func (t *TemplateConvertor) apiFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s string) (valLst []string, val interface{}, err error) {
	if err = checkTemplateDetail(&temp); err != nil {
		return nil, nil, err
	}
	// http://RESOURCE-JAVA-SERVICE/api/ala/getCircuitNo?resid=${k_RESID}:outPara
	reg, _ := regexp.Compile(`://(.*?)/`)

	var uri, GetValjsonPath string // 最终取值的jsonPath
	ctx.Infof("origin url %s", temp.MapExpr)
	uriLst := strings.Split(temp.MapExpr, "::")
	if len(uriLst) == 2 {
		uri = uriLst[0]
		GetValjsonPath = uriLst[1]
	} else {
		uri = temp.MapExpr
		GetValjsonPath = "data"
	}

	lst := reg.FindStringSubmatch(uri)
	if len(lst) < 2 {
		return nil, nil, fmt.Errorf("invalid api mapExpr %s", temp.MapExpr)
	}
	serviceName := lst[1]
	var addr string // 最终地址
	if !strings.Contains(serviceName, ".") &&
		!strings.Contains(serviceName, ":") &&
		!strings.Contains(serviceName, "localhost") {
		//说明是nacos 服务名称

		ins, err := cache.GetNacosCli().Instance(serviceName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get ip and port from nacos failed")
		}
		addr = fmt.Sprintf("%s:%d", ins.IP, ins.Port)
	} else {
		addr = serviceName
	}
	uri = strings.Replace(uri, serviceName, addr, 1)

	valLst = getVariables(uri)
	for _, v := range valLst {

		v_, err := t.GetAttr(v, msg, s)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get val failed, key %s, ", v)
		}
		s_ := cast.ToString(v_)
		uri = strings.ReplaceAll(uri, fmt.Sprintf("${%s}", v), s_)
	}
	uri, err = util.UrlEncode(uri)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "url encode failed, %s", uri)
	}
	ctx.Infof("api Func url: %s", uri)

	req, err := http.NewRequestWithContext(ctx.Context, "GET", uri, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "apiFunc create request failed")

	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trace_id", ctx.TraceVal)

	response, err := util.HttpDo(req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "apiFunc do http request failed")
	}
	defer func() {
		_ = response.Body.Close()
	}()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, errors.Wrapf(err,
			"apiFunc parse request body failed, method=get,uri=%s, body=%s", uri, string(b))
	}
	ctx.Debugf("api Func res: %s", string(b))

	if gjson.GetBytes(b, "code").String() != "0000" {
		return nil, nil, fmt.Errorf("apiFunc code of response is not '0000',body: %s", string(b))
	}

	ctx.Infof("getVal by jsonPath: %s", GetValjsonPath)
	ss := gjson.GetBytes(b, GetValjsonPath)
	var ssVal string
	if !ss.Exists() {
		GetValjsonPath = "data." + GetValjsonPath
		ctx.Infof("getVal by jsonPath: %s", GetValjsonPath)
		ss = gjson.GetBytes(b, GetValjsonPath)
		if !ss.Exists() {
			ctx.Infoln("getVal by jsonPath failed %s", GetValjsonPath)
			return
		}
		ssVal = ss.String()
	} else {
		ssVal = ss.String()
	}
	ssVal = strings.Trim(ssVal, " ")
	ctx.Infof("get val %s", ssVal)

	return valLst, ssVal, nil
}

// RegCaseFunc
// regCase-正则判断与截取：此种情况下，映射表达式格式为多个条件表达式判断，多个条件之间“;”分隔。表达式操作符为“match”，表示正则匹配。
// 例如“{attr1} match {regexprA} {attr2,regexprB};{attr1} match {regexprA} {attr2,regexprB}”。
// 执行逻辑：以“;”拆解表达式为子表达式，遍历子表达式，若srcField为attr1的告警属性若匹配正则regexprA，
// 则映射以正则作用于srcField为attr2的告警属性后得到的值；若不匹配则继续下一个子表达式，若匹配上则退出遍历。
func (t *TemplateConvertor) RegCaseFunc(ctx kctx.Ctx, v model.Elem, msg model.IMsg, s string) (interface{}, error) {
	// {title} match {端口down} {alarm_node,.*:(.*)};{title} match {BGP连接中断} {bussiness_name,(.*)}
	// {title} match {端口down} {alarm_node,.*:(.*)};{title} match {.*} {bussiness_name,(.*)}
	var val string
	cases := strings.Split(v.MapExpr, ";")
	for _, c := range cases {
		// {title} match {端口down}
		attr1, expr1, attr2, expr2, e := parseRegCase(c)
		if e != nil {
			ctx.Errorf("caseFunc failed, case expr: %s", c)
			continue
		}
		vi, err := t.GetAttr(attr1, msg, s)
		if err != nil {
			ctx.Errorf("reg caseFunc: %v", err)
			continue
		}
		ss := util.ConvertToString(vi)

		m, e := regexp.Match(expr1, []byte(ss))
		if e != nil {
			ctx.Errorf("reg caseFunc, match err, expr: %s, data: %q, err: %v", expr1, ss, e)
			continue
		}
		if !m {
			ctx.Debugf("reg caseFunc not match expr: %s, data: %q, ", expr1, ss)
			continue
		}
		vi2, err := t.GetAttr(attr2, msg, s)
		if err != nil {
			ctx.Errorf("%s caseFunc failed, attr not exist, attr  %s, err: %v", attr2, err)
			continue
		}
		ss2 := util.ConvertToString(vi2)
		var reg *regexp.Regexp
		expr2 = strings.Trim(expr2, "'")
		reg, err = regexp.Compile(expr2)
		if err != nil {
			return nil, fmt.Errorf("caseFunc failed, compile rexpr failed, expr %q  err: %v", expr2, err)
		}
		out := reg.FindAllStringSubmatch(ss2, -1)
		for _, v__ := range out {
			for i := 1; i < len(v__); i++ {
				val += " " + v__[i]
			}
		}
		val = strings.Trim(val, " ")
		return val, nil
	}
	return nil, fmt.Errorf("caseFunc not found")
}

type FindAttrF func(resType ResType, attrName string) (string, interface{}, error)

func FindResAttrFunc(alarmResClass string, dev map[string]string, port map[string]string) FindAttrF {
	// alarmResClass 告警资源类型
	return func(resType ResType, attrName string) (ResClass string, val interface{}, err error) {

		switch resType {
		case ResTypeDev:
			v, ok := dev[attrName]
			if ok {
				return alarmResClass, v, nil
			}
			return alarmResClass, nil, fmt.Errorf("%s not exist in dev info", attrName)
		default:
			v, ok := port[attrName]
			if ok {
				return alarmResClass, v, nil
			}
			return alarmResClass, nil, fmt.Errorf("%s not exist in %v info", attrName, resType)
		}
	}
}

// FindAttrFunc
// findAttr-资源属性查找：此种情况下，映射表达式格式为“资源类别|属性条件|目标属性”，其中属性条件可以是多个，多个以逗号分隔；
// 目标属性只能是一个。例如“PORT|attr1=${XXX},attr2=${YYY}|attr5”，
// 系统根据条件属性attr1,attr2在端口资源缓存中定位到attr5（即attr1=XXX&&attr2=YYY作为条件，找到attr5的属性值）。
// 若存在多个满足条件的目标属性，则返回第一个。
func (t *TemplateConvertor) FindAttrFunc(ctx kctx.Ctx, v model.Elem, f FindAttrF) (string, interface{}, error) {
	// PORT|attr1
	l := strings.Split(v.MapExpr, "|")
	if len(l) != 2 {
		return "", nil, fmt.Errorf("invalid map expr %s for findAttr", v.MapExpr)
	}
	resType := ResType(strings.Trim(l[0], " "))
	attr := strings.Trim(l[1], " ")

	return f(resType, attr)
}

func (t *TemplateConvertor) GetAttr(key string, m model.IMsg, s string) (interface{}, error) {
	ss, err := getValInJson(s, key)

	if err != nil {
		v, ok := m.Get(key)
		if !ok {
			return nil, errors.Wrapf(err, "key %s not exist in msg and result, err: %v", key, err)
		}
		return v, nil
	}
	return ss, nil

}

func parseRegCase(s string) (attr1, expr1, attr2, expr2 string, err error) {
	// {title} match {端口down} {alarm_node,.*:(.*)}
	lst := strings.Split(s, "match")
	if len(lst) != 2 {
		err = fmt.Errorf("invalidate case formate %q", s)
		return
	}
	attr1 = delSpace(lst[0])
	ll := strings.Split(lst[1], "}")
	if len(ll) == 0 {
		err = fmt.Errorf("invalidate case formate %q", s)
		return
	}

	expr1 = delSpace(ll[0])
	lst2 := strings.Split(s, "{")
	if len(lst2) == 0 {
		err = fmt.Errorf("invalidate case formate %q", s)
		return
	}
	l := delSpace(lst2[len(lst2)-1])
	lst3 := strings.Split(l, ",")
	if len(lst3) != 2 {
		return
	}
	attr2 = delSpace(lst3[0])
	expr2 = delSpace(lst3[1])
	return
}
func delSpace(s string) string {
	s = strings.Trim(s, " ")
	s = strings.Trim(s, "{")
	s = strings.Trim(s, "}")
	s = strings.Trim(s, "/")
	return s
}

// DictFunc
/*
   {
     "srcField": "k_ALARMTYPE",
     "dstField": "SNC_TITLE",
     "alarmType": "DS0001,DS0002",
     "mapType": "dict",
     "mapExpr": "dict_1"
   },

*/
func (t *TemplateConvertor) DictFunc(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, dic model.Dicts, s string) (interface{}, error) {
	val, err := t.GetAttr(temp.SrcField, msg, s)
	if err != nil {
		return nil, fmt.Errorf("get field val failed, key %s", temp.SrcField)
	}

	for _, v := range dic {
		if v.DictCode == temp.MapExpr {
			for _, dLst := range v.DictList {
				if len(dLst) != 2 {
					continue
				}
				v_ := util.ConvertToString(val)
				if dLst[0] == v_ {
					return dLst[1], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("not found val in Rule dict, dictCode %s, val %v", temp.MapExpr, val)
}

type ApiCacheOption struct {
	// CacheKey
	/*示例：
	    func(varLst []string) string {
			// unios-alarm:api: 默认
			k := config.GetString("Redis.Key.ApiCacheKey")
			k2 := strings.Join(varLst, ":")
			return fmt.Sprintf("%s%s", k, k2)
		}
	*/
	CacheKey   func(varLst []string) string // 缓存key 的算法  返回缓存的key 值
	ExpireHour int                          // 超时时间 小时
}

// ApiCache
//  apiCache-API调用并缓存：此种情况下，同api方式，但是调用后会进行缓存，其中key=unios-alarm:api:{para1}:{para2}，
// 其中para1、para2为api调用入参，有几个就拼接几个。缓存失效时间为24小时。
func (t *TemplateConvertor) ApiCache(ctx kctx.Ctx, temp model.Elem, msg model.IMsg, s string) (interface{}, error) {
	varLst := getVariables(temp.MapExpr)
	if len(varLst) == 0 {
		return nil, fmt.Errorf("invalid mapexpr in ApiCache, variable is empty %s", temp.MapExpr)
	}
	// 获取不是变量的参数
	val2 := getParamVal(temp.MapExpr)

	valLst := make([]string, 0)
	for _, v := range varLst {
		v_, err := t.GetAttr(v, msg, s)
		if err != nil {
			ctx.Error(err)
			continue
		}

		valLst = append(valLst, util.ConvertToString(v_))
	}
	valLst = append(valLst, val2...)
	lst := strings.Split(temp.MapExpr, "::")
	if len(lst) > 0 {
		valLst = append(valLst, lst[1:len(lst)]...)
	}
	sort.Slice(valLst, func(i, j int) bool {
		return valLst[i] < valLst[j]
	})

	key := t.Option.ApiCache.CacheKey(valLst)

	ss, err := t.cache.GetString(ctx, key)
	if err == nil {
		ctx.Infof("get cache success,key: %q val: %q", key, ss)
		return ss, nil
	}
	ctx.Errorf("get cache failed,key: %s get info from apiFunc, err: %v", key, err)

	val, err := t.APIFunc(ctx, temp, msg, s)
	if err != nil {
		return nil, errors.Wrapf(err, "api func failed, err: %v", err)
	}
	if val == nil {
		return nil, fmt.Errorf("get call api return empty")
	}
	if err = t.cache.SetStringWithExpire(ctx, key, util.ConvertToString(val), t.ExpireHour*3600); err != nil {
		return nil, errors.Wrapf(err, "set info to cache failed, err: %v", err)
	}
	ctx.Infof("apiCache success key %q val %q", key, val)
	return val, nil
}

// getParamVal
// rpc/res/alarm/getOnuOpticalSplitter?oltdevid=${k_RESID}&ponport=${k_RESPARA}&onuid=${k_LOCATEINFO}&ostype=1
// 获取非变量的参数
func getParamVal(expr string) []string {
	lst := strings.Split(expr, "?")
	if len(lst) < 2 {
		return nil
	}
	suffix := strings.Join(lst[1:], "?")
	lst2 := strings.Split(suffix, "&")
	out := make([]string, 0)
	for _, v := range lst2 {
		if strings.Contains(v, "${") {
			continue
		}
		i := strings.Index(v, "=")
		if len(v) == i+1 {
			continue
		}

		valStr := v[i+1:]
		if valStr == "" {
			continue
		}
		i2 := strings.Index(valStr, "::")
		if i2 > 0 {
			valStr = valStr[:i2]
		}

		out = append(out, valStr)
	}

	return out
}

func (t *TemplateConvertor) AutoSwitchFunc(ctx kctx.Ctx, v model.Elem, m model.IMsg, s *string) (interface{}, error) {
	var val interface{}
	if v.DstField == "AlarmTypeID" {
		v1 := m.GetStr(v.SrcField)
		i3eCheckSum2 := crc32.ChecksumIEEE([]byte(v1))
		typeId := "OMC" + strconv.Itoa(int(i3eCheckSum2))
		val = typeId
		_, errq := t.cache.HGetString(ctx, "unios-alarm:alarmtype", typeId)
		if errq != nil {
			alarmTypeCode, err3 := t.getAlarmType(ctx, v1)
			if alarmTypeCode != "" {
				val = alarmTypeCode
			}
			if err3 != nil {
				//仍然未找到,新增
				result := gjson.Get(*s, "Severity")
				ctx.Debugf("目前文本%s, %v", *s)
				if result.Exists() {
					body := map[string]interface{}{
						"alarmTypeCode":  typeId,
						"alarmTypeDescr": v1,
						"alarmClassCode": "DS",
						"alarmLevel":     result.String(),
						"cleanTtl":       0,
						"specialty":      "ALL",
						"cleanType":      "",
						"status":         1,
					}
					if err := t.apiAddAlarmTypeDetail(ctx, body); err != nil {
						ctx.Warnf("apiAddAlarmTypeDetail failed, err: %v", err)
					}
				}
			}
		}
	}
	if val == nil {
		return nil, fmt.Errorf("not found val in Rule autoSwitch, MapExpr %s", v.MapExpr)
	}
	return val, nil
}

func (t *TemplateConvertor) getAlarmType(ctx kctx.Ctx, typeDesc string) (string, error) {
	//serviceName := "RESOURCE-JAVA-SERVICE"
	serviceName := config.GetString("Pub.Alarm.Service")
	//说明是nacos 服务名称
	ins, err := cache.GetNacosCli().Instance(serviceName)
	if err != nil {
		return "", errors.Wrapf(err, "get ip and port from nacos failed")
	}

	uri := fmt.Sprintf("http://%s:%d/%s",
		ins.IP,
		ins.Port,
		"api/ala/alarmpalne/getalarmtype")

	body := map[string]interface{}{
		"alarmTypeDesc": typeDesc,
	}
	b, _ := json.Marshal(body)
	ctx.Infof("getalarmtype url %s , body %s", uri, string(b))
	req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.Wrapf(err, "new request failed")
	}

	resp, err := util.HttpDoWithResByte(ctx, req)
	if err != nil {
		return "", errors.Wrapf(err, "http do failed")
	}
	ctx.Debugf("getAlarmType body %s", string(resp))

	type AlarmType struct {
		Code string `json:"code"`
		Data []struct {
			AlarmClassCode string `json:"alarmClassCode"`
			AlarmTypeCode  string `json:"alarmTypeCode"`
			AlarmTypeDescr string `json:"alarmTypeDescr"`
			Alarmlevel     int    `json:"alarmlevel"`
			AuthFlag       string `json:"authFlag"`
			AuthUser       string `json:"authUser"`
			CreateTime     string `json:"createTime"`
			CreateUser     string `json:"createUser"`
			ModifyTime     string `json:"modifyTime"`
			Specialty      string `json:"specialty"`
			SpecialtyDesc  string `json:"specialtyDesc"`
			Status         int    `json:"status"`
		} `json:"data"`
		Msg string `json:"msg"`
		Tip string `json:"tip"`
	}
	alarmType := AlarmType{}
	if err := json.Unmarshal(resp, &alarmType); err != nil {
		return "", errors.Wrapf(err, "unmarshal failed")
	}
	if len(alarmType.Data) > 0 {
		return alarmType.Data[0].AlarmTypeCode, nil
	}
	return "", fmt.Errorf("not found alarmTypeCode")
}

func (t *TemplateConvertor) apiAddAlarmTypeDetail(ctx kctx.Ctx, body map[string]interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Errorf("addalarmtypedetail Error %s", string(debug.Stack()))
		}
	}()

	//serviceName := "RESOURCE-JAVA-SERVICE"
	serviceName := config.GetString("Pub.Alarm.Service")
	//说明是nacos 服务名称
	ins, err := cache.GetNacosCli().Instance(serviceName)
	if err != nil {
		return errors.Wrapf(err, "get ip and port from nacos failed")
	}

	uri := fmt.Sprintf("http://%s:%d/%s",
		ins.IP,
		ins.Port,
		"api/ala/alarmpalne/addalarmtypedetail")

	b, _ := json.Marshal(body)
	ctx.Infof("url %s, body %s", uri, string(b))
	req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return errors.Wrapf(err, "new request failed")
	}
	resp, err := util.HttpDoWithResByte(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "http do failed")
	}
	ctx.Debugf("addalarmtypedetail body %s", string(resp))
	return nil
}
