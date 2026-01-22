package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	esapiv8 "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/db/es"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

type IESSender interface {
	BatchSend(ctx kctx.Ctx, index string, infos []EsSendCap) (map[string]interface{}, error)
	Send(ctx kctx.Ctx, index string, id string, info interface{}) error
}

type ESErrCallback func(ctx kctx.Ctx, index string, infos []EsSendCap, esErr error, res map[string]interface{})

func EsSendRepoInit(ctx kctx.Ctx, getter IConfigGetter, prefix string, f ESErrCallback) (*EsSendRepo, error) {
	return NewEsRepo(
		ctx,
		getter,
		es.Config{
			Addresses: strings.Split(getter.GetString(prefix+".Urls"), ","),
			Username:  getter.GetString(prefix + ".Username"),
			Password:  getter.GetString(prefix + ".Password"),
			Version:   getter.GetInt(prefix + ".Version"),
		},
		getter.GetString("Pub.Es.Index.Prefix"),
		getter.GetString(prefix+".Index"),
		getter.GetString(prefix+".IndexLayout"),
		getter.GetString(prefix+".Type"),
		getter.GetInt(prefix+".WaitTime"),
		make(chan EsSendCap),
		f,
	)
}

// NewEsRepoWith 使用getter
/*
Urls = "http://localhost:9200"
# ES 类型 ， 6.几版本时， 需要配置, 7.几时，配置为空
Type = ""
Username = ""
Password = ""
#时区根据现场ES 情况配置 UTC   Asia/Shanghai
Zone = "Asia/Shanghai"
# 生成入ES语句 等待PUB_ES_WAITTIME秒入单位秒， 如WatiTime为5时， 表示每5s， 如果有数据则入es一次
WaitTime = 1
# $index_${!now().format_timestamp("20060102", "${Alarm.Es.Zone}")}
Index = "am_ala_alarmsummary"
IndexLayout = "20060102"
*/
func NewEsRepoWith(ctx kctx.Ctx, getter IConfigGetter,
	prefix string, msgChan chan EsSendCap, errCallback ESErrCallback) (*EsSendRepo, error) {
	username := prefix + ".Username"
	password := prefix + ".Password"
	urls := prefix + ".Urls"
	index := prefix + ".Index"
	indexLayout := prefix + ".IndexLayout"
	estype := prefix + ".Type"
	waitTime := prefix + ".WaitTime"
	apiKey := prefix + ".ApiKey"
	version := prefix + ".Version"

	cli, err := NewEsRepo(
		ctx,
		getter,
		es.Config{
			Addresses: strings.Split(getter.GetString(urls), ","),
			Username:  getter.GetString(username),
			Password:  getter.GetString(password),
			APIKey:    getter.GetString(apiKey),
			Version:   getter.GetInt(version),
		},
		getter.GetString("Pub.ES.Index.Prefix"),
		getter.GetString(index),
		getter.GetString(indexLayout),
		getter.GetString(estype),
		getter.GetInt(waitTime),
		msgChan,
		errCallback)
	if err != nil {
		return nil, errors.Wrapf(err, "NewEsRepo")
	}
	cli.GroupPrefix = prefix

	return cli.SyncNacos(), nil
}

func NewEsRepo(
	ctx kctx.Ctx,
	getter IConfigGetter,
	conf es.Config,
	prefix,
	index,
	indexTimeLayout,
	esType string,
	waitTime int,
	msgChan chan EsSendCap,
	f ESErrCallback,
) (*EsSendRepo, error) {

	var cf context.CancelFunc
	ctx.Context, cf = context.WithCancel(ctx.Context)
	if len(conf.Addresses) > 0 && strings.Contains(conf.Addresses[0], ",") {
		lst := gset.NewStrSetFrom(strings.Split(conf.Addresses[0], ","))
		conf.Addresses = lst.Slice()
	}

	cli, err := es.NewClient(&conf)
	if err != nil {
		cf()
		return nil, errors.Wrapf(err, "Init ES failed, err: %v", err)
	}
	s := &EsSendRepo{
		prefix:             prefix,
		ctx:                ctx,
		getter:             getter,
		index:              index,
		indexTimeLayout:    indexTimeLayout,
		cli:                cli,
		esType:             esType,
		intervalSecond:     waitTime,
		msgChan:            msgChan,
		esErrCallback:      f,
		sendIntervalSecond: waitTime,
		Close: func() error {
			cf()
			return nil
		},
	}
	return s, nil
}

type EsSendRepo struct {
	ctx                kctx.Ctx
	prefix             string // indexPrefix
	GroupPrefix        string // 配置文件中的组前缀
	getter             IConfigGetter
	index              string
	indexKey           string // 配置文件中的key, 用于nacos 热更新
	indexTimeLayoutKey string //
	indexTimeLayout    string // 索引时间最后后缀， 时间layout 用于nacos 热更新
	cli                *es.Client
	esType             string         // 6.几  _doc  7.几， 为空
	intervalSecond     int            // 批量插入时间间隔
	msgChan            chan EsSendCap // 批量插入的时候的消息channel, nacos 热更新的时候， 得保存channel 的一致性
	esErrCallback      ESErrCallback  // es 插入异常时候的回调
	sendIntervalSecond int
	Close              func() error
}

// SyncNacos nacos 热更新
func (e *EsSendRepo) SyncNacos() (cli *EsSendRepo) {

	if e.getter.IsNacosSync() {
		prefix := e.GroupPrefix
		username := prefix + ".Username"
		password := prefix + ".Password"
		urls := prefix + ".Urls"
		var index, indexLayout string
		if e.indexKey != "" {
			index = e.index
		} else {
			index = prefix + ".Index"
		}
		if e.indexTimeLayoutKey != "" {
			e.indexTimeLayout = e.indexTimeLayoutKey
		} else {
			indexLayout = prefix + ".IndexLayout"
		}
		estype := prefix + ".Type"

		waitTime := prefix + ".WaitTime"

		changed := make(chan struct{})

		lst := []string{username, password, urls, index, indexLayout, estype, waitTime}
		for _, v := range lst {
			func(v string) {
				if err := ListenNacos(e.getter, v, func(namespace, group, dataId, data string) {
					e.getter.Set(v, data)
					e.ctx.Infof("es config change %s  %s", v, data)
					changed <- struct{}{}

				}); err != nil {
					e.ctx.Errorf("listen nacos config failed: %v", err)
				}
			}(v)
		}
		go func() {
			for {
				select {
				case <-e.ctx.Done():
					return
				case _, ok := <-changed:
					if !ok {
						return
					}

					cli_, err := NewEsRepoWith(e.ctx, e.getter, e.prefix, e.msgChan, e.esErrCallback)
					if err != nil {
						e.ctx.Errorf("NewEsRepoWith failed: %v", err)
						continue
					}
					cli_.SetIndexKey(e.indexKey)
					cli_.SetIndexTimeLayoutKey(e.indexTimeLayoutKey)

					if err := cli_.Ping(); err != nil {
						e.ctx.Errorf("es cli ping failed: %v, config: %+v", err, cli.cli)
						continue
					}
					// 成功了
					cf, err := cli_.Run()
					if err != nil {
						e.ctx.Errorf("es cli run failed: %v", err)
						continue
					}

					cli_.Close = func() error {
						cf()
						return nil
					}

					// 关闭原有的

					_ = e.Close()
					// 赋值新的
					cli = cli_
				}
			}
		}()
		cli = e
		return
	}

	return e
}

func (e *EsSendRepo) Ping() error {
	m, err := e.cli.Info()
	if err != nil {
		return errors.Wrapf(err, "Ping")
	}

	if m["tagline"] == "You Know, for Search" {
		e.ctx.Infof("%#v", m)
		return nil
	}

	return fmt.Errorf("esCli.Info() failed: %#v", m)
}

func (e *EsSendRepo) Run() (cf func(), err error) {
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(e.sendIntervalSecond))
		defer ticker.Stop()

		temp := make(map[string][]EsSendCap, 0)
		lock := sync.Mutex{}
		for {
			select {
			case <-e.ctx.Done():
				e.ctx.Infof("EsSendRepo Closed")
				return
			case msg, ok := <-e.msgChan:
				if !ok {
					return
				}
				lock.Lock()
				temp[msg.Index] = append(temp[msg.Index], msg)
				lock.Unlock()
			case <-ticker.C:
				if len(temp) > 0 {
					ctx := kctx.NewContext(context.TODO(), util.NewUUIDStr())
					lock.Lock()
					for index, datalist := range temp {
						res, esErr := e.BatchSend(ctx, index, datalist)
						if esErr != nil {
							if e.esErrCallback != nil {
								ctx.Infof("index %s write to es failed, callback success", index)
								go func(datalist_ []EsSendCap) {
									e.esErrCallback(ctx, index, datalist_, esErr, res)
								}(datalist)
							}
							ctx.Infof("write to es failed, err: %v", esErr)
						} else {
							var isFalse bool
							f, ok := res["errors"]
							if ok {
								t, ok := f.(bool)
								if ok && t {
									if lst, ok := res["items"].([]interface{}); ok {
										if len(lst) > 5 {
											res["items"] = lst[:4]
										}
									}
									ctx.Infof("write to es failed, err: %#v", res)
									isFalse = true
								}
							}
							if !isFalse {
								if lst, ok := res["items"].([]interface{}); ok {
									if len(lst) > 5 {
										res["lenOfItems"] = len(lst)
										res["items"] = lst[:4]
									}
								}

								ctx.Infof("write to es success %v", res)
							}
						}
					}
					lock.Unlock()
					temp = make(map[string][]EsSendCap, 0)
				}
			}
		}
	}()

	return nil, nil
}

func (e *EsSendRepo) SetIndex(i string) {
	e.index = i
}
func (e *EsSendRepo) SetIndexTimeLayout(v string) {
	e.indexTimeLayout = v
}
func (e *EsSendRepo) SetIndexKey(i string) {
	e.index = e.getter.GetString(i)
	e.indexKey = i
}
func (e *EsSendRepo) SetIndexTimeLayoutKey(i string) {
	e.indexTimeLayout = e.getter.GetString(i)
	e.indexTimeLayoutKey = i
}

func (e *EsSendRepo) GetCli() *es.Client {
	return e.cli
}

// BatchSend 批量直接发送
func (e *EsSendRepo) BatchSend(ctx kctx.Ctx, index string, infos []EsSendCap) (map[string]interface{}, error) {
	return e.batchSendWith(ctx, index, infos)
}

func (e *EsSendRepo) batchSendWith(ctx kctx.Ctx, index string, infos []EsSendCap) (map[string]interface{}, error) {
	ctx.Debugf("index %s in es Send len %d", index, len(infos))

	info := make(map[string]string, len(infos))
	for _, v := range infos {
		if v.Id == "" {
			v.Id = util.NewUUIDStr()
		}
		info[v.Id] = v.Content
	}
	//ctx.Info("es batch write contain trace_id %v", tids.Slice())

	//resBody, err := e.cli.BulkInsert(index, docs, e.esType)
	//resBody, err := e.cli.BulkInsertWithContent(ctx, index, docs, e.esType)

	resBody, err := e.cli.UpsertWithString(ctx, es.ReqArg{
		Index:   index,
		DocType: e.esType,
		Refresh: "true",
	}, info)
	//resBody, err := e.cli.BulkInsertAndRefresh(index, docs, e.esType)
	if err != nil {
		return resBody, errors.Wrapf(err, "ES buildInsert failed")
	}

	return resBody, nil

}

func (e *EsSendRepo) BatchSendWith(ctx kctx.Ctx, index string, infos []EsSendCap) (map[string]interface{}, error) {
	return e.batchSendWith(ctx, index, infos)
}

type EsSendCap struct {
	TraceId string `json:"trace_id"`
	Index   string `json:"index"`
	Id      string `json:"id"`      //
	Content string `json:"content"` // 插入es的string
}

// Send
// 此处使用upsert , 所以需要传入id
func (e *EsSendRepo) Send(ctx kctx.Ctx, index string, id string, info interface{}) error {
	b, err := json.Marshal(info)
	if err != nil {
		return errors.Wrapf(err, "marshal es send info failed")
	}

	bb := string(b)
	if bb == "" {
		return nil
	}

	e.msgChan <- EsSendCap{
		TraceId: ctx.TraceVal,
		Index:   index,
		Id:      id,
		Content: bb,
	}

	return nil
}

// QueryWith 查询某字段的特定值是否存在。 如查询k_sumalarmid=aaaa 的记录是否存在, 如果存在， 则返回数据结构体
/* res 返回的数据结构
{
    "_shards": {
        "failed": 0,
        "skipped": 0,
        "successful": 1,
        "total": 1
    },
    "hits": {
        "hits": [
            {
                "_id": "RmyZQokB-xgmc8dtsD2d",
                "_ignored": [
                    "k_RECOVERMSG.keyword",
                    "s_SUMMARY.keyword",
                    "k_RAWMSG.keyword"
                ],
                "_index": "am_ala_alarmsummary__20230711",
                "_score": 0.2876821,
                "_source": {
                    "k_AGENT_RULE": "NE40E_simulator",
                    "k_RESPROPCODE": "INTER",
                    "k_RESPROPNAME": "互联电路",

                    "t_STARTTIME": "2023-05-29 15:38:00"
                },
                "_type": "_doc"
            }
        ],
        "max_score": 0.2876821,
        "total": {
            "relation": "eq",
            "value": 1
        }
    },
    "timed_out": false,
    "took": 2
}
*/
/*
返回不存在的时候
{
    "_shards": {
        "failed": 0,
        "skipped": 0,
        "successful": 1,
        "total": 1
    },
    "hits": {
        "hits": [],
        "max_score": null,
        "total": {
            "relation": "eq",
            "value": 0
        }
    },
    "timed_out": false,
    "took": 3
}
*/
func (e *EsSendRepo) QueryWith(ctx kctx.Ctx, key, val string) (gjson.Result, error) {

	query := fmt.Sprintf(`
{
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "%s": {
              "value": "%s"
            }
          }
        }
      ]
    }
  }
}`, key, val)
	var g gjson.Result
	index := fmt.Sprintf("%s%s_*", e.prefix, e.index)
	m, err := e.cli.Search(index, query, 1)
	if err != nil {
		return g, err
	}
	b, _ := json.Marshal(m)
	g = gjson.ParseBytes(b)
	return g, nil
}

// QueryByIdAndIndexLst 在指定的索引中查询某id是否存在 如果存在则返回
func (e *EsSendRepo) QueryByIdAndIndexLst(ctx kctx.Ctx, id string, indexList []string) (gjson.Result, error) {
	wg := sync.WaitGroup{}
	var out gjson.Result

	for _, v := range indexList {
		if v == "" {
			continue
		}
		wg.Add(1)
		go func(index string) {
			defer wg.Done()
			g, err := e.QueryById(ctx, index, id)
			if err != nil {
				return
			}
			if g.Get("found").Bool() {

				out = g
			}
		}(v)
	}

	wg.Wait()
	return out, nil
}
func (e *EsSendRepo) QueryById(ctx kctx.Ctx, index, id string) (gjson.Result, error) {
	if e.cli.GetVersion() == 8 {
		cli := e.cli.RawClientV8()
		// 根据id查询
		req := esapiv8.GetRequest{
			Index:      index,
			DocumentID: id,
		}
		// 发送请求并处理响应
		res, err := req.Do(ctx, cli)
		if err != nil {
			return gjson.Result{}, errors.Wrapf(err, "es query faild")
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return gjson.ParseBytes(b), nil
	}

	cli := e.cli.RawClient()
	// 根据id查询
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}
	// 发送请求并处理响应
	res, err := req.Do(ctx, cli)
	if err != nil {
		return gjson.Result{}, errors.Wrapf(err, "es query faild")
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	return gjson.ParseBytes(b), nil
}

// IsExist 判断是否存在
func (e *EsSendRepo) IsExist(ctx kctx.Ctx, key, val string) (isExist bool, id, index string, err error) {
	g, err := e.QueryWith(ctx, key, val)
	if err != nil {
		return false, "", "", errors.Wrap(err, "query with")
	}
	id = g.Get("hits.hits.0._id").String()
	index = g.Get("hits.hits.0._index").String()

	return g.Get("hits.total.value").Int() >= 1, id, index, nil
}

// Update 部分更新
func (e *EsSendRepo) Update(ctx kctx.Ctx, id, index string, jsonInfo json.RawMessage) ([]byte, error) {
	body, err := sjson.SetRawBytes([]byte(""), "doc", jsonInfo)
	if err != nil {
		return nil, errors.Wrap(err, "set es update body")
	}
	if e.cli.GetVersion() == 8 {
		res, err := e.cli.RawClientV8().Update(index, id, bytes.NewReader(body), func(request *esapiv8.UpdateRequest) {
			request.Refresh = "true"
		})
		if err != nil {
			return nil, errors.Wrapf(err, "es update")
		}

		defer res.Body.Close()

		b, _ := io.ReadAll(res.Body)
		return b, nil
	}

	res, err := e.cli.RawClient().Update(index, id, bytes.NewReader(body), func(request *esapi.UpdateRequest) {
		request.Refresh = "true"
	})
	if err != nil {
		return nil, errors.Wrapf(err, "es update")
	}

	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	/*
		{
		    "_index": "am_ala_alarmsummary__20230711",
		    "_type": "_doc",
		    "_id": "RmyZQokB-xgmc8dtsD2d",
		    "_version": 2,
		    "result": "updated",
		    "forced_refresh": true,
		    "_shards": {
		        "total": 2,
		        "successful": 1,
		        "failed": 0
		    },
		    "_seq_no": 1,
		    "_primary_term": 1
		}
	*/
	return b, nil
}
