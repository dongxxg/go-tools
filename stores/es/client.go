// 基于ElasticSearch官方SDK的封装，支持V7.x版本。
// godoc: https://pkg.go.dev/github.com/elastic/go-elasticsearch
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dongxxg/go-tools/lib/mapping"
	"github.com/dongxxg/go-tools/log"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type Client struct {
	cfg    *Config
	client *es.Client
}

// NewClient根据配置新建Client
func NewClient(cfg *Config) (*Client, error) {
	log.Infof("creating es client, go-elastic version: %s", es.Version)
	c, err := es.NewClient(*cfg.getESConfig())
	if err != nil {
		return nil, err
	}

	return &Client{
		client: c,
		cfg:    cfg,
	}, nil
}

// 返回go-elasticsearch的Client对象指针
// 如果封装Client没有覆盖到需要的方法，可以直接用原始Client
func (c *Client) RawClient() *es.Client {
	return c.client
}

// 查询cluster信息
func (c *Client) Info() (map[string]interface{}, error) {
	res, err := c.client.Info()
	if err != nil {
		log.Errorf("Info: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("Info: error parsing the response body: %s", err)
		return nil, err
	}

	if res.IsError() {
		err = fmt.Errorf("[%s] Info: status error", res.Status())
	}

	return r, err
}

// Perform 代理HTTP请求，可以通过类似postman的方式执行ES命令
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	return c.client.Perform(req)
}

// SQLQuery 执行SQL查询
func (c *Client) SQLQuery(sql string) (map[string]interface{}, error) {
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

// Get返回一个文档
func (c *Client) Get(index, id string) (map[string]interface{}, error) {
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

// Index添加或者更新文档记录。新增文档时，id为可选参数。
func (c *Client) Index(index, body, id, indextype string) (map[string]interface{}, error) {
	res, err := c.client.Index(
		index,                              // Index name
		strings.NewReader(body),            // Document body
		c.client.Index.WithDocumentID(id),  // Document ID
		c.client.Index.WithRefresh("true"), // Refresh
		c.client.Index.WithDocumentType(indextype),
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

// Search搜索文档。
// query是查询条件`{"query" : { "match" : { "title" : "test" } }}`，size是返回hits数
func (c *Client) Search(index, query string, size int) (map[string]interface{}, error) {
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

// 创建索引
func (c *Client) CreateIndex(index string, shardNum, replicasNum int) (map[string]interface{}, error) {
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

// DeleteIndex删除索引
func (c *Client) DeleteIndex(index string) (map[string]interface{}, error) {
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

// BulkInsert批量在一个Index中插入文档。
// 批量添加时，增加指定数据的类型，数据的类型可为raw、_doc等
// 官方文档：https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
func (c *Client) BulkInsert(index string, docs []string, esType string) (map[string]interface{}, error) {
	return c.bulkInsert(index, docs, esType, false)
}

// 插入es并立即同步到不同分片
func (c *Client) BulkInsertAndRefresh(index string, docs []string, esType string) (map[string]interface{}, error) {
	return c.bulkInsert(index, docs, esType, true)
}
func (c *Client) bulkInsert(index string, docs []string, esType string, refresh bool) (map[string]interface{}, error) {
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
		log.Errorf("BulkInsert: ES error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Errorf("BulkInsert error parsing the response body: %s", err)
		return nil, err
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