package es

// 如何将search结果转为struct，见Test文件

type SearchHit struct {
	Score  *float64               `json:"_score"`  // computed score
	Index  string                 `json:"_index"`  // index name
	ID     string                 `json:"_id"`     // external or internal
	Type   string                 `json:"_type"`   // type
	Source map[string]interface{} `json:"_source"` // stored document source
}

// SearchHits specifies the list of search hits.
type SearchHits struct {
	MaxScore *float64     `json:"max_score"` // maximum score of all hits
	Hits     []*SearchHit `json:"hits"`      // the actual hits returned
}

// SearchResult is the result of a search in Elasticsearch.
type SearchResult struct {
	TookInMillis float64     `json:"took"`            // search time in milliseconds
	Hits         *SearchHits `json:"hits"`            // the actual search hits
	TimedOut     bool        `json:"timed_out"`       // true if the search timed out
	Error        string      `json:"error,omitempty"` // used in MultiSearch only
}
