package receiver

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
	model2 "unitechs.com/cno/kit/lib/mq/model"
	"unitechs.com/cno/kit/template/model"
	"unitechs.com/cno/kit/util"
)

type IBreakTime interface {
	GetBreakTime() (time.Time, error)
	SetBreakTime(t string) error
	Delete() error
}

var repo IBreakTime

type ReceiverRest struct {
	p       model.ParamRest
	ctx     kctx.Ctx
	context context.Context
	cf      context.CancelFunc
	period  int
}

var restPool *ants.Pool

func NewReceiverRestWith(ctx kctx.Ctx, p model.ParamRest, poolSize int, r IBreakTime) (*ReceiverRest, error) {
	var period int
	if p.PubRestPeriod == "0" {
		period = 10
	} else {
		n, err := strconv.Atoi(p.PubRestPeriod)
		if err != nil {
			ctx.Errorf("invalid rest period formate %s, %v", p.PubRestPeriod, err)
			n = 10
		}
		period = n
	}
	var err error
	restPool, err = ants.NewPool(poolSize)
	if err != nil {
		return nil, errors.Wrap(err, "init pool failed")
	}

	repo = r

	return &ReceiverRest{
		p:      p,
		ctx:    ctx,
		period: period,
	}, nil

}
func (r *ReceiverRest) run() {
	p := r.p
	for {
		select {
		case <-r.context.Done():
			r.ctx.Infof("rest down success")
			return
		default:

		}
		breakTime, err := repo.GetBreakTime()
		if err != nil {
			r.ctx.Errorf("get redis breakTime failed: %v", err)
			time.Sleep(time.Minute * time.Duration(r.period))
			continue
		}
		r.ctx.Infof("breakTime %s", breakTime.Format(util.TimeFormatDefault))

		endTime := time.Now().Add(-time.Minute * 2)
		r.ctx.Infof("endTime %s", endTime.Format(util.TimeFormatDefault))
		// 最少获取10s 的数据
		//if breakTime.After(endTime.Add(-time.Second * 10)) {
		//	time.Sleep(time.Minute * time.Duration(r.period))
		//	continue
		//}

		switch p.PubRestPage {
		case "y", "Y", "YES", "Yes", "yes":
			// 使用分页
			RestQueryWithPage(r.ctx, p, breakTime, endTime)
		default:
			RestQueryWithoutPage(r.ctx, p, breakTime, endTime)
		}
		if err = repo.SetBreakTime(endTime.Format(util.TimeFormatDefault)); err != nil {
			r.ctx.Errorf("set redis breakTime failed: %v", err)
			time.Sleep(time.Minute * time.Duration(r.period))
			continue
		}

		r.ctx.Infof("time: %s, interval: %d min,  waiting for next period... ...",
			time.Now().Format(util.TimeFormatDefault), r.period)
		ticker := time.NewTicker(time.Minute * time.Duration(r.period))
		select {
		case <-r.context.Done():
			return
		case <-ticker.C:
			r.ctx.Infof("rest down success")
		}
	}
}

func (r *ReceiverRest) Collect(ruleId string, param map[string]interface{}, f model2.HandleFunc) (cf func(), err error) {
	r.context, r.cf = context.WithCancel(r.ctx.Context)
	cf = func() {}

	p := model.ParamRest{}
	if err = util.MarshalWith(param, &p); err != nil {
		err = errors.Wrapf(err, "failed to parse rest param, param: %+v", param)
		return
	}
	r.ctx.Debugf("rest param %+v", param)
	if err = util.MarshalWith(param, &p); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal rest param,  param: %+v", param)
		return
	}
	r.p = p

	go func() {
		for v := range MsgChan {
			select {
			case <-r.context.Done():
				return
			default:

			}

			if err = restPool.Submit(func() {
				c := context.WithValue(context.TODO(), KeyRuleID, ruleId)
				c = context.WithValue(c, KeyRuleType, model.RuleTypeRest)
				ctx := kctx.NewContext(c, util.NewUUIDStr())
				f(ctx, model2.ConsumerMsg{
					Timestamp: time.Now(),
					Value:     []byte(v),
				})

			}); err != nil {
				r.ctx.Errorf("submit task to pool failed in rest, err: %v", err)
			}
		}
	}()

	go func() {
		r.run()
	}()

	return func() {
		r.Close()
	}, nil
}
func (r *ReceiverRest) Close() error {
	r.cf()
	return nil
}

// MsgChan 消息通道， rest 获取消息后，放入此通道， 消费者消费此通道数据至ants 池
var MsgChan chan string

func init() {
	MsgChan = make(chan string, 1)
}

const MaxErrNum = 50

// RestQueryWithPage 带分页的请求
func RestQueryWithPage(ctx kctx.Ctx, p model.ParamRest, breakTime, endTime time.Time) {
	var currentPage int64 = 1
	errNum := 0
	for {
		data, curPage, totalPage, err := RestQuery(ctx, currentPage, p, breakTime, endTime)
		if err != nil {
			err = errors.WithMessagef(err, "CollectRest")
			ctx.Errorf("rest query failed, err: %v", err)
			errNum += 1
			if errNum > MaxErrNum {
				return
			}
			time.Sleep(time.Second)
			continue
		}
		if totalPage == 0 && currentPage == 0 {
			errNum += 1
			continue
		}
		data.ForEach(func(key, value gjson.Result) bool {
			MsgChan <- value.String()
			return true
		})
		if totalPage != 0 && currentPage != 0 && curPage == totalPage {
			return
		}

		currentPage += 1
		if currentPage > totalPage {
			return
		}

		// 防止并发过大
		time.Sleep(time.Second)
	}
}

func RestQueryWithoutPage(ctx kctx.Ctx, p model.ParamRest, breakTime time.Time, endTime time.Time) {
	data, _, _, err := RestQuery(ctx, 1, p, breakTime, endTime)
	if err != nil {
		ctx.Errorf("[CollectRest.RestQuery] failed, err=%v", err)
	} else {
		data.ForEach(func(key, value gjson.Result) bool {
			MsgChan <- value.String()
			return true
		})
	}
}

// RestQuery rest 接口查询
func RestQuery(ctx kctx.Ctx, currentPage int64, p model.ParamRest, bt, endTime time.Time) (msg gjson.Result, curPage, totalPage int64, err error) {

	//bt, err := repo.GetBreakTime()
	//if err != nil {
	//	return msg, 0, 0, errors.Wrap(err, "get break time in redis failed")
	//}
	if currentPage == 0 {
		currentPage = 1
	}
	endTimeStr := endTime.Format(util.TimeFormatDefault)

	u := url.Values{}

	resPara := make(map[string]string)
	if err = json.Unmarshal([]byte(p.PubRestParam), &resPara); err != nil {
		ctx.Errorf("unmarshal pubRestParam failed: %v", err)
	}
	for k, v := range resPara {
		u.Add(k, v)
	}

	//u.Add("orgSeverity", "1,2,3,4")
	//u.Add("networkTypeTop", "3")
	//u.Add("activeStatus", "1")
	//u.Add("objectClass", "900")
	//u.Add("pageSize", "1000")

	u.Add("beginTime", bt.Format(util.TimeFormatDefault))
	u.Add("endTime", endTimeStr)
	u.Add("current", strconv.Itoa(int(currentPage)))

	s := u.Encode()

	if p.PubRestMethod == "" {
		p.PubRestMethod = "POST"
	}
	p.PubRestMethod = strings.ToUpper(p.PubRestMethod)

	p.PubRestUrl += "?" + s
	ctx.Infof("uri: %s", p.PubRestUrl)
	req, err := http.NewRequest(p.PubRestMethod, p.PubRestUrl, nil)
	if err != nil {
		return msg, 0, 0, errors.WithMessage(err, "NewRequest failed get rest request")
	}
	headerMap := make(map[string]string)

	if p.PubRestHeader != "" {
		if err := json.Unmarshal([]byte(p.PubRestHeader), &headerMap); err != nil {
			return msg, 0, 0, errors.WithMessage(err, "config unmarshal header to map failed")
		}
	}

	headerMap["app-id"] = p.PubRestAppid
	timestamp := time.Now().UnixNano() / 1e6
	headerMap["timestamp"] = strconv.FormatInt(timestamp, 10)
	sign := computeSign(p.PubRestAppid, p.PubRestAppSecret, timestamp)
	headerMap["sign"] = sign
	headerMap["Content-type"] = "application/x-www-form-urlencoded" // 表单的header
	for k, v := range headerMap {
		req.Header.Set(k, v)
	}
	b, _ := json.Marshal(headerMap)

	ctx.Infof("header: %s", string(b))

	res, err := http.DefaultClient.Do(req)
	//res, err := http.PostForm(uri, u)
	if err != nil {
		return msg, 0, 0, errors.WithMessagef(err, "rest query failed")
	}
	defer res.Body.Close()

	bb, _ := io.ReadAll(res.Body)
	g := gjson.ParseBytes(bb)
	if g.Get(p.PubRestStatusField).String() != p.PubRestStatusValue {
		return msg, 0, 0, fmt.Errorf("rest query get failed, invalidate code, successCode=%s, currCode=%s, res=%s",
			p.PubRestStatusValue, g.Get(p.PubRestStatusField).String(), string(bb))
	}
	curPage = g.Get(p.PubRestCurPage).Int()

	totalPage = g.Get(p.PubRestPagesizeField).Int()

	// 设置当前时间为下一个breakTime
	ctx.Infof("[RestQueryWithPage] total page %d, current page: %d, endTime: %s",
		totalPage,
		currentPage,
		endTime,
	)
	r := g.Get(p.PubRestDataField)
	if r.Exists() {
		return r, curPage, totalPage, nil
	}
	return msg, 0, 0, fmt.Errorf("invalid result, data field %s not exist  %s",
		p.PubRestDataField, g.String(),
	)
}

// computeSign 通过配置文件计算 sign 值
// appid,timestamp,appSecret  转换为16位大写的md5码
func computeSign(appId, appSecret string, timestamp int64) string {
	s := fmt.Sprintf("%s,%d,%s", appId, timestamp, appSecret)
	res := GetMD5Encode(s)
	return strings.ToUpper(res)
}

func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
