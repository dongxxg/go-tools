package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"strings"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
	sd "unitechs.com/unios-dice/uni-base/core/sd/interface"
	"unitechs.com/unios-dice/uni-base/core/sd/nacos"
)

type IsEncode string

const (
	IsEncodeY IsEncode = "y"
	IsEncodeN          = "n"
)

var Base64Encode *BaseEncode

type BaseEncode struct {
	m sync.Map
}

// GetPassword 密码转码
func GetPassword(ctx kctx.Ctx, pwd, key string) string {
	return Base64Encode.GetPassword(ctx, pwd, key)
}

func NewBaseEncode(lst SubscribeDictLst) {
	b := &BaseEncode{}
	for _, v := range lst {
		b.m.Store(v.ParaCode, v.IsEncode)
	}
	Base64Encode = b
}

// GetPassword 密码转码
func (i *BaseEncode) GetPassword(ctx kctx.Ctx, pwd, key string) string {
	if i.IsBase64Encode(key) {
		p, err := ParsePwd(pwd)
		if err != nil {
			ctx.Errorf("%v", err)
			return pwd
		}
		return p
	}
	return pwd
}

func (i *BaseEncode) Set(key string, v IsEncode) {
	i.m.Store(key, v)
}

func (i *BaseEncode) IsBase64Encode(key string) bool {
	if i == nil {
		return true
	}
	if d, ok := i.m.Load(key); ok {
		if s, ok := d.(IsEncode); ok {
			switch IsEncode(strings.ToLower(string(s))) {
			case IsEncodeY:
				return true
			}
		}
	}

	return false
}

func ParsePwd(s string) (string, error) {
	p, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.WithMessagef(err, "invalidate password, must base64 encrypted")
	}
	return string(p), nil
}

type Base64EncodeMap map[string]IsEncode

type ParaCode string

var nacosCli sd.Client

func GetNacosCli() sd.Client {
	if nacosCli == nil {
		nacosCli = nacos.NewClient()
	}
	return nacosCli
}

type SubscribeDictLst []SubscribeDictElem
type SubscribeDictElem struct {
	IsEncode IsEncode `json:"isEncode"`
	ParaCode string   `json:"paraCode"`
}

// GetSubscribeDictPara  读取3.7.3.订阅参数字典值查询 接口并存为缓存
func GetSubscribeDictPara(ctx kctx.Ctx, serviceName string) (SubscribeDictLst, error) {
	//Service = "ALARM-JAVA-SERVICE"
	if serviceName == "" {
		return nil, errors.New("invalidate Pub.Alarm.Service name, is empty")
	}

	ins, err := GetNacosCli().Instance(serviceName)
	if err != nil {
		return nil, errors.WithMessagef(err, "nacos get instance by serviceName[%s] failed", serviceName)
	}
	// 192.168.6.87:9777
	// curl http://192.168.6.87:9777/api/ala/rule/getSubscribeDictPara
	uri := fmt.Sprintf("http://%s:%d/api/ala/rule/getSubscribeDictPara", ins.IP, ins.Port)
	ctx.Debugf("get subscribe dict uri: %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.WithMessagef(err, "new request failed")
	}
	res, err := util.HttpDo(req)
	if err != nil {
		return nil, errors.WithMessagef(err, "http request failed")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.WithMessagef(err, "parse res failed")
	}
	g := gjson.ParseBytes(b)
	if g.Get("code").String() != "0000" {
		return nil, fmt.Errorf("res code invalidate, res: %s", string(b))
	}
	out := SubscribeDictLst{}

	if err := json.Unmarshal([]byte(g.Get("data.paraList").String()), &out); err != nil {
		return nil, errors.WithMessagef(err, "parse result failed")
	}
	ctx.Debugf("get subscribe dict success: %s", string(b))
	return out, nil
}
