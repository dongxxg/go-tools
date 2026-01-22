package es

// 基于ElasticSearch官方SDK的封装，支持V7.x版本。
// godoc: https://pkg.go.dev/github.com/elastic/go-elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
	"io"
	"net/http"
	"strings"
	"unitechs.com/cno/kit/kctx"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	v8es "github.com/elastic/go-elasticsearch/v8"
	v8esapi "github.com/elastic/go-elasticsearch/v8/esapi"
	"unitechs.com/unios-dice/uni-base/core/log"
	"unitechs.com/unios-dice/uni-base/lib/mapping"
)

type Client struct {
	cfg                *Config
	client             *es.Client
	clientV8           *v8es.Client
	sendIntervalSecond int // 间隔多少秒发送一次
	version            int
}

// NewClient 根据配置新建Client
func NewClient(cfg *Config) (*Client, error) {
	log.Infof("creating es client, go-elastic version: %s config %v", es.Version, *cfg.getESConfig())
	if cfg.Version == 8 {
		c, err := v8es.NewClient(*cfg.getV8ESConfig())
		if err != nil {
			return nil, err
		}
		cli := &Client{
			version:  8,
			clientV8: c,
			cfg:      cfg,
		}
		m, err := cli.clientV8.Ping()
		if err != nil {
			return nil, errors.Wrapf(err, "esClient info failed")
		}
		fmt.Println(m)
		return cli, nil

	} else {
		c, err := es.NewClient(*cfg.getESConfig())
		if err != nil {
			return nil, err
		}
		cli := &Client{
			version: 7,
			client:  c,
			cfg:     cfg,
		}
		m, err := cli.client.Ping()
		if err != nil {
			return nil, errors.Wrapf(err, "esClient info failed")
		}
		fmt.Println(m)
		return cli, nil
	}
}
func (c *Client) Config() *Config {
	return c.cfg
}

func (c *Client) GetVersion() int {
	return c.version
}

// RawClient 返回go-elasticsearch的Client对象指针
// 如果封装Client没有覆盖到需要的方法，可以直接用原始Client
func (c *Client) RawClient() *es.Client {
	return c.client
}
func (c *Client) RawClientV8() *v8es.Client {
	return c.clientV8
}

// Info 查询cluster信息
func (c *Client) Info() (map[string]interface{}, error) {

	var r map[string]interface{}
	var err error
	if c.version == 8 {
		res, err := c.clientV8.Info()
		if err != nil {
			log.Errorf("Info: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("Info: error parsing the response body: %s", err)
			return nil, err
		}
		if res.IsError() {
			err = fmt.Errorf("[%s] Info: status error", res.Status())
		}

	} else {
		res, err := c.client.Info()
		if err != nil {
			log.Errorf("Info: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("Info: error parsing the response body: %s", err)
			return nil, err
		}
		if res.IsError() {
			err = fmt.Errorf("[%s] Info: status error", res.Status())
		}
	}

	return r, err
}

// Perform 代理HTTP请求，可以通过类似postman的方式执行ES命令
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	if c.version == 8 {
		return c.clientV8.Perform(req)
	}

	return c.client.Perform(req)
}

// SQLQuery 执行SQL查询
func (c *Client) SQLQuery(sql string) (map[string]interface{}, error) {
	if c.version == 8 {
		req := v8esapi.SQLQueryRequest{
			Body:       strings.NewReader(`{"query":"` + sql + `"}`),
			Pretty:     true,
			Human:      true,
			ErrorTrace: true,
		}
		res, err := req.Do(context.Background(), c.client)
		if err != nil {
			log.Errorf("SQLQuery: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("SQLQuery: error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] SQLQuery: error status, sql=%s", res.Status(), sql)
		}
		return r, err
	}

	req := esapi.SQLQueryRequest{
		Body:       strings.NewReader(`{"query":"` + sql + `"}`),
		Pretty:     true,
		Human:      true,
		ErrorTrace: true,
	}
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		log.Errorf("SQLQuery: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("SQLQuery: error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] SQLQuery: error status, sql=%s", res.Status(), sql)
	}

	return r, err
}

// Get 返回一个文档
func (c *Client) Get(index, id string) (map[string]interface{}, error) {
	if c.version == 8 {
		res, err := c.clientV8.Get(index, id)
		if err != nil {
			log.Errorf("Get: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("Get error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] Error Get document ID=%s index=%s", res.Status(), id, index)
		}

		return r, err
	}

	res, err := c.client.Get(index, id)
	if err != nil {
		log.Errorf("Get: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("Get error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] Error Get document ID=%s index=%s", res.Status(), id, index)
	}

	return r, err
}

// Index 添加或者更新文档记录。新增文档时，id为可选参数。
func (c *Client) Index(index, body, id, indextype string) (map[string]interface{}, error) {
	if c.version == 8 {
		res, err := c.clientV8.Index(
			index,                                // Index name
			strings.NewReader(body),              // Document body
			c.clientV8.Index.WithDocumentID(id),  // Document ID
			c.clientV8.Index.WithRefresh("true"), // Refresh
			//c.client.Index.WithDocumentType(indextype),
		)
		if err != nil {
			log.Errorf("Index: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("Index error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] Error indexing document ID=%s index=%s", res.Status(), id, index)
			log.Debugf("Error indexing document: %s", mapping.Repr(res))
		}

		return r, err
	}

	res, err := c.client.Index(
		index,                              // Index name
		strings.NewReader(body),            // Document body
		c.client.Index.WithDocumentID(id),  // Document ID
		c.client.Index.WithRefresh("true"), // Refresh
		//c.client.Index.WithDocumentType(indextype),
	)
	if err != nil {
		log.Errorf("Index: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("Index error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] Error indexing document ID=%s index=%s", res.Status(), id, index)
		log.Debugf("Error indexing document: %s", mapping.Repr(res))
	}

	return r, err
}

// Search 搜索文档。
// query是查询条件`{"query" : { "match" : { "title" : "test" } }}`，size是返回hits数
func (c *Client) Search(index, query string, size int) (map[string]interface{}, error) {
	if c.version == 8 {
		res, err := c.clientV8.Search(
			c.clientV8.Search.WithContext(context.Background()),
			c.clientV8.Search.WithIndex(index),
			c.clientV8.Search.WithBody(strings.NewReader(query)),
			c.clientV8.Search.WithTrackTotalHits(true),
			c.clientV8.Search.WithPretty(),
			c.clientV8.Search.WithSize(size),
		)
		if err != nil {
			log.Errorf("ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("Search error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			// Print the response status and error information.
			log.Errorf("[%s] %s: %s",
				res.Status(),
				r["error"].(map[string]interface{})["type"],
				r["error"].(map[string]interface{})["reason"],
			)

			return nil, fmt.Errorf("[%s] Error Search document query=%s index=%s", res.Status(), query, index)
		}

		// Print the response status, number of results, and request duration.
		log.Debugf(
			"search %s with query %s - [%s] %d hits; took: %dms",
			index, query, res.Status(),
			len(r["hits"].(map[string]interface{})["hits"].([]interface{})), // ES 6和7的total格式不一样，这里取数组长度
			int(r["took"].(float64)),
		)
		// Print the ID and document source for each hit.
		// for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		// 	log.Tracef(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		// }
		return r, nil
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(context.Background()),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(strings.NewReader(query)),
		c.client.Search.WithTrackTotalHits(true),
		c.client.Search.WithPretty(),
		c.client.Search.WithSize(size),
	)
	if err != nil {
		log.Errorf("ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("Search error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		// Print the response status and error information.
		log.Errorf("[%s] %s: %s",
			res.Status(),
			r["error"].(map[string]interface{})["type"],
			r["error"].(map[string]interface{})["reason"],
		)

		return nil, fmt.Errorf("[%s] Error Search document query=%s index=%s", res.Status(), query, index)
	}

	// Print the response status, number of results, and request duration.
	log.Debugf(
		"search %s with query %s - [%s] %d hits; took: %dms",
		index, query, res.Status(),
		len(r["hits"].(map[string]interface{})["hits"].([]interface{})), // ES 6和7的total格式不一样，这里取数组长度
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	// for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
	// 	log.Tracef(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	// }
	return r, nil
}

// DeleteDoc 从index删除一个文档
func (c *Client) DeleteDoc(index, id string) (map[string]interface{}, error) {
	if c.version == 8 {
		res, err := c.clientV8.Delete(index, id)
		if err != nil {
			log.Errorf("DeleteDoc: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("DeleteDoc error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error DeleteDoc index=%s, id=%s", res.Status(), index, id)
		}

		return r, err
	}

	res, err := c.client.Delete(index, id)
	if err != nil {
		log.Errorf("DeleteDoc: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("DeleteDoc error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error DeleteDoc index=%s, id=%s", res.Status(), index, id)
	}

	return r, err
}

// DeleteDocByQuery 根据从index删除文档
func (c *Client) DeleteDocByQuery(index, query string) (map[string]interface{}, error) {
	if c.version == 8 {

		res, err := c.clientV8.DeleteByQuery([]string{index}, strings.NewReader(query))
		if err != nil {
			log.Errorf("DeleteDocByQuery: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("DeleteDocByQuery error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error DeleteDocByQuery index=%s, query=%s", res.Status(), index, query)
		}

		return r, err
	}

	res, err := c.client.DeleteByQuery([]string{index}, strings.NewReader(query))
	if err != nil {
		log.Errorf("DeleteDocByQuery: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("DeleteDocByQuery error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error DeleteDocByQuery index=%s, query=%s", res.Status(), index, query)
	}

	return r, err
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(index string, shardNum, replicasNum int) (map[string]interface{}, error) {
	if c.version == 8 {

		res, err := c.clientV8.Indices.Create(
			index,
			c.clientV8.Indices.Create.WithBody(strings.NewReader(fmt.Sprintf(`{
			"settings": {
				"number_of_shards": %d,
				"number_of_replicas": %d
			}
		}`, shardNum, replicasNum))),
		)
		if err != nil {
			log.Errorf("CreateIndex: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("CreateIndex error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error CreateIndex index=%s", res.Status(), index)
		}

		return r, err
	}

	res, err := c.client.Indices.Create(
		index,
		c.client.Indices.Create.WithBody(strings.NewReader(fmt.Sprintf(`{
			"settings": {
				"number_of_shards": %d,
				"number_of_replicas": %d
			}
		}`, shardNum, replicasNum))),
	)
	if err != nil {
		log.Errorf("CreateIndex: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("CreateIndex error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error CreateIndex index=%s", res.Status(), index)
	}

	return r, err
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(index string) (map[string]interface{}, error) {
	if c.version == 8 {
		res, err := c.clientV8.Indices.Delete([]string{index})
		if err != nil {
			log.Errorf("DeleteIndex: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Errorf("DeleteIndex error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error DeleteIndex index=%s", res.Status(), index)
		}

		return r, err
	}

	res, err := c.client.Indices.Delete([]string{index})
	if err != nil {
		log.Errorf("DeleteIndex: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("DeleteIndex error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error DeleteIndex index=%s", res.Status(), index)
	}

	return r, err
}

// BulkInsert 批量在一个Index中插入文档。
// 批量添加时，增加指定数据的类型，数据的类型可为raw、_doc等
// 官方文档：https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
func (c *Client) BulkInsert(index string, docs []string, esType string) (map[string]interface{}, error) {
	return c.bulkInsert(index, docs, esType, false)
}

// BulkInsertAndRefresh 插入es并立即同步到不同分片
func (c *Client) BulkInsertAndRefresh(index string, docs []string, esType string) (map[string]interface{}, error) {
	return c.bulkInsert(index, docs, esType, true)
}
func (c *Client) BulkInsertWithContent(ctx kctx.Ctx, index string, docs []string, esType string) (map[string]interface{}, error) {
	return c.bulkInsertWithContext(ctx, index, docs, esType, true)
}

func (c *Client) bulkInsertWithContext(ctx kctx.Ctx, index string, docs []string, esType string, refresh bool) (map[string]interface{}, error) {
	var buffer bytes.Buffer
	// 拼接body格式如下
	// { "index" : { "_index" : "test", "_id" : "1" } }
	// { "field1" : "value1" }
	// { "delete" : { "_index" : "test", "_id" : "2" } }
	// { "create" : { "_index" : "test", "_id" : "3" } }
	// { "field1" : "value3" }
	// { "update" : {"_id" : "1", "_index" : "test"} }
	// { "doc" : {"field2" : "value2"} }
	for _, doc := range docs {
		// 每个doc必须在一行中，中间不能出现换行
		doc = strings.Replace(doc, "\n", "", -1)

		buffer.WriteString(`{"index":{}}`)
		buffer.WriteString("\n")
		buffer.WriteString(doc)
		buffer.WriteString("\n")
	}

	if c.version == 8 {
		req := v8esapi.BulkRequest{
			Index:        index,
			Body:         strings.NewReader(buffer.String()),
			DocumentType: esType,
		}

		if refresh {
			req.Refresh = "true"
		}
		res, err := req.Do(context.Background(), c.clientV8)
		if err != nil {
			ctx.Errorf("BulkInsert: ES error: %s", err)
			return nil, err
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			ctx.Errorf("BulkInsert error parsing the response body: %s", err)
			return nil, err
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error BulkInsert index=%s, res: %v", res.Status(), index, r)
		}

		if errFlag, ok := r["errors"].(bool); ok && errFlag {
			// 这种情况下在返回的第一个参数中有详细信息
			err = fmt.Errorf("error BulkInsert index=%s: one or more of the operations in the bulk request did not complete successfully %v", index, r)
		}

		return r, err

	}
	req := esapi.BulkRequest{
		Index:        index,
		Body:         strings.NewReader(buffer.String()),
		DocumentType: esType,
	}

	if refresh {
		req.Refresh = "true"
	}
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		ctx.Errorf("BulkInsert: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		ctx.Errorf("BulkInsert error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error BulkInsert index=%s, res: %v", res.Status(), index, r)
	}

	if errFlag, ok := r["errors"].(bool); ok && errFlag {
		// 这种情况下在返回的第一个参数中有详细信息
		err = fmt.Errorf("error BulkInsert index=%s: one or more of the operations in the bulk request did not complete successfully %v", index, r)
	}

	return r, err

}

func (c *Client) bulkInsert(index string, docs []string, esType string, refresh bool) (map[string]interface{}, error) {
	var buffer bytes.Buffer
	// 拼接body格式如下
	// { "index" : { "_index" : "test", "_id" : "1" } }
	// { "field1" : "value1" }
	// { "delete" : { "_index" : "test", "_id" : "2" } }
	// { "create" : { "_index" : "test", "_id" : "3" } }
	// { "field1" : "value3" }
	// { "update" : {"_id" : "1", "_index" : "test"} }q
	// { "doc" : {"field2" : "value2"} }
	for _, doc := range docs {
		// 每个doc必须在一行中，中间不能出现换行
		doc = strings.Replace(doc, "\n", "", -1)

		buffer.WriteString(`{"index":{}}`)
		buffer.WriteString("\n")
		buffer.WriteString(doc)
		buffer.WriteString("\n")
	}
	if c.version == 8 {
		req := v8esapi.BulkRequest{
			Index:        index,
			Body:         strings.NewReader(buffer.String()),
			DocumentType: esType,
		}
		if refresh {
			req.Refresh = "true"
		}
		res, err := req.Do(context.Background(), c.clientV8)
		if err != nil {
			return nil, errors.Wrapf(err, "req.Do")
		}
		defer res.Body.Close()

		var r map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
			return nil, errors.Wrapf(err, "decode")
		}

		if res.IsError() {
			err = fmt.Errorf("[%s] error BulkInsert index=%s", res.Status(), index)
		}

		if errFlag, ok := r["errors"].(bool); ok && errFlag {
			// 这种情况下在返回的第一个参数中有详细信息
			err = fmt.Errorf("error BulkInsert index=%s: one or more of the operations in the bulk request did not complete successfully", index)
		}

		return r, err

	}
	req := esapi.BulkRequest{
		Index:        index,
		Body:         strings.NewReader(buffer.String()),
		DocumentType: esType,
	}
	if refresh {
		req.Refresh = "true"
	}
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		return nil, errors.Wrapf(err, "req.Do")
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, errors.Wrapf(err, "decode")
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] error BulkInsert index=%s", res.Status(), index)
	}

	if errFlag, ok := r["errors"].(bool); ok && errFlag {
		// 这种情况下在返回的第一个参数中有详细信息
		err = fmt.Errorf("error BulkInsert index=%s: one or more of the operations in the bulk request did not complete successfully", index)
	}

	return r, err
}

type ReqArg struct {
	Index   string
	DocType string
	Refresh string
}

// Upsert
// POST /_bulk
// { "update" : {"_id" : "bb","_type":"_doc","_index": "am_ala_alarmsummary_20230828"} }
// { "doc" : {"name" : "cc"}, "doc_as_upsert" : true }
func (c *Client) Upsert(ctx kctx.Ctx, reqArg ReqArg, content map[string]map[string]interface{}) (map[string]interface{}, error) {
	var bulkRequests []string
	var buffer bytes.Buffer
	if reqArg.Index == "" {
		return nil, fmt.Errorf("index can`t be empty")
	}
	// Loop through your documents and create bulk requests
	for id, v := range content {
		// Create a bulk request with an update action
		if reqArg.DocType == "" {
			buffer.WriteString(fmt.Sprintf(`{ "update" : {"_id" : "%s", "_index" : "%s"} }`,
				id,
				reqArg.Index,
			))
		} else {
			buffer.WriteString(fmt.Sprintf(`{ "update" : {"_id" : "%s", "_index" : "%s","_type":"%s"} }`,
				id,
				reqArg.Index,
				reqArg.DocType,
			))
		}

		// Add the update request to the bulkRequests slice
		doc := `{"doc_as_upsert":true}`
		temp := "{}"
		for k, v2 := range v {
			temp, _ = sjson.Set(temp, k, v2)
		}
		doc, _ = sjson.SetRaw(doc, "doc", temp)
		buffer.WriteString("\n")
		buffer.WriteString(doc)
		buffer.WriteString("\n")
		bulkRequests = append(bulkRequests, buffer.String())
	}
	// Join all the bulk requests into a single string
	bulkRequestBody := strings.Join(bulkRequests, "\n")
	if c.version == 8 {
		// Create a bulk request
		req := v8esapi.BulkRequest{
			Index:        reqArg.Index,
			Body:         strings.NewReader(bulkRequestBody),
			DocumentType: reqArg.DocType,
			Refresh:      reqArg.Refresh,
		}

		// Perform the bulk request
		res, err := req.Do(ctx, c.clientV8)
		if err != nil {
			return nil, errors.Wrapf(err, "req.Do, body %s", bulkRequestBody)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)

		r := make(map[string]interface{})
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, errors.Wrapf(err, "decode %s", string(b))
		}

		if res.IsError() {
			return nil, errors.Wrapf(fmt.Errorf("%v", r), "res is error")
		}

		return r, nil
	}
	// Create a bulk request
	req := esapi.BulkRequest{
		Index:        reqArg.Index,
		Body:         strings.NewReader(bulkRequestBody),
		DocumentType: reqArg.DocType,
		Refresh:      reqArg.Refresh,
	}

	// Perform the bulk request
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, errors.Wrapf(err, "req.Do, body %s", bulkRequestBody)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)

	r := make(map[string]interface{})
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, errors.Wrapf(err, "decode %s", string(b))
	}

	if res.IsError() {
		return nil, errors.Wrapf(fmt.Errorf("%v", r), "res is error")
	}

	return r, nil
}

// UpsertWithString
// POST /_bulk
// { "update" : {"_id" : "bb","_type":"_doc","_index": "am_ala_alarmsummary_20230828"} }
// { "doc" : {"name" : "cc"}, "doc_as_upsert" : true }
func (c *Client) UpsertWithString(ctx kctx.Ctx, reqArg ReqArg, content map[string]string) (map[string]interface{}, error) {
	//var bulkRequests []string
	var buffer bytes.Buffer
	if reqArg.Index == "" {
		return nil, fmt.Errorf("index can`t be empty")
	}
	// Loop through your documents and create bulk requests
	for id, v := range content {
		if id == "" {
			continue
		}
		// Create a bulk request with an update action
		if reqArg.DocType == "" {
			buffer.WriteString(fmt.Sprintf(`{ "update" : {"_id" : "%s", "_index" : "%s"} }`,
				id,
				reqArg.Index,
			))
		} else {
			buffer.WriteString(fmt.Sprintf(`{ "update" : {"_id" : "%s", "_index" : "%s","_type":"%s"} }`,
				id,
				reqArg.Index,
				reqArg.DocType,
			))
		}

		// Add the update request to the bulkRequests slice
		doc := `{"doc_as_upsert":true}`
		doc, _ = sjson.SetRaw(doc, "doc", v)
		buffer.WriteString("\n")
		buffer.WriteString(doc)
		buffer.WriteString("\n")
		buffer.WriteString("\n")
		//bulkRequests = append(bulkRequests, buffer.String())
	}
	// Join all the bulk requests into a single string
	//bulkRequestBody := strings.Join(bulkRequests, "\n")
	// Create a bulk request
	if c.version == 8 {

		req := v8esapi.BulkRequest{
			Index:        reqArg.Index,
			Body:         strings.NewReader(buffer.String()),
			DocumentType: reqArg.DocType,
			Refresh:      reqArg.Refresh,
		}

		// Perform the bulk request
		res, err := req.Do(ctx, c.clientV8)
		if err != nil {
			return nil, errors.Wrapf(err, "req.Do, body")
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		r := make(map[string]interface{})
		if err = json.Unmarshal(b, &r); err != nil {
			return nil, errors.Wrapf(err, "decode %s", string(b))
		}

		if res.IsError() {
			return nil, errors.Wrapf(fmt.Errorf("%v", r), "res is error")
		}

		return r, nil

	}

	req := esapi.BulkRequest{
		Index:        reqArg.Index,
		Body:         strings.NewReader(buffer.String()),
		DocumentType: reqArg.DocType,
		Refresh:      reqArg.Refresh,
	}

	// Perform the bulk request
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, errors.Wrapf(err, "req.Do, body")
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	r := make(map[string]interface{})
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, errors.Wrapf(err, "decode %s", string(b))
	}

	if res.IsError() {
		return nil, errors.Wrapf(fmt.Errorf("%v", r), "res is error")
	}

	return r, nil

}
