package es

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"unitechs.com/unios-dice/uni-base/lib/mapping"
	"unitechs.com/unios-dice/uni-base/lib/stringx"
)

var (
	ES_HOST = "http://127.0.0.1:9200"
)

func TestInfo(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	info, err := c.Info()
	assert.NoError(t, err)

	t.Log(info)
}

func TestPerform(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", "_cat/indices?v", nil)
	assert.NoError(t, err)

	res, err := c.client.Perform(req)
	assert.NoError(t, err)

	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	t.Log(ioutil.ReadAll(res.Body))
}

func TestSQLQuery(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.SQLQuery("select name from customer")
	assert.NoError(t, err)

	t.Log(res)
}

func TestIndex(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.Index("customer", "{\"name\":\"abc\"}", "", "_doc")
	assert.NoError(t, err)
	// fmt.Println(mapping.Repr(res))

	t.Log(res)
}

func TestGet(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.Get("customer", "1")
	assert.NoError(t, err)

	t.Log(res)
}

func TestSearch(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.Search("customer", `{"query" : { "match" : { "name" : "abc" } }}`, 50)
	assert.NoError(t, err)
	// fmt.Println(mapping.Repr(res))

	t.Log(res)
}

func TestDeleteDoc(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.DeleteDoc("customer", "100")
	// assert.NoError(t, err)

	t.Log(res)
}

func DeleteDocByQuery(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.DeleteDocByQuery("customer", `{
		"query": {
			"match": {
			"customer.id": "1"
			}
		}
		}`)
	assert.NoError(t, err)

	t.Log(res)
}

func TestCreateIndex(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.CreateIndex("index2", 1, 1)
	assert.NoError(t, err)

	t.Log(res)
}

func TestDeleteIndex(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.DeleteIndex("index2")
	assert.NoError(t, err)

	t.Log(res)
}

func TestBulkInsert(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	// docs := []string{}
	// docs = append(docs, fmt.Sprintf(`{"name":"%s"}`, stringx.Rand()), fmt.Sprintf(`{"name":"%s"}`, stringx.Rand()))

	docs := []string{}
	for i := 0; i < 1000; i++ {
		docs = append(docs, fmt.Sprintf(`{"name":"%s"}`, stringx.Rand()), fmt.Sprintf(`{"name":"%s"}`, stringx.Rand()))
	}

	esType := "_doc"
	res, err := c.BulkInsert("customer", docs, esType)
	assert.NoError(t, err)

	t.Log(res)
}

// 测试将search结果转为struct
func TestUnmarshalResult(t *testing.T) {
	c, err := NewClient(NewConfig([]string{ES_HOST}))
	assert.NoError(t, err)

	res, err := c.Search("customer", `{"query" : { "match" : { "name" : "abc" } }}`, 10)
	assert.NoError(t, err)

	unMarshaler := mapping.NewUnmarshaler("json")

	fina := SearchResult{}
	err = unMarshaler.Unmarshal(res, &fina)
	assert.NoError(t, err)
}
