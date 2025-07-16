package http_util

import (
	"bytes"
	"io"
	"time"
)

const (
	POST = "POST"
	GET  = "GET"
)

const (
	ApplicationJSON  = "application/json"
	ApplicationForm  = "application/x-www-form-urlencoded"
	ApplicationOctet = "application/octet-stream"
)

const (
	defMethod      = GET              // 默认请求方式
	defContentType = ApplicationJSON  // 默认Context-Type
	defTimeout     = 20 * time.Second // 默认超时时间
)

type RequestParam struct {
	Method      string
	ContentType string
	Timeout     time.Duration
	Body        []byte
	Query       map[string]string
	header      map[string]string
}

func defaultRequest() *RequestParam {
	return &RequestParam{
		Method:      defMethod,
		ContentType: defContentType,
		Timeout:     defTimeout,
		Body:        nil,
		header:      make(map[string]string),
	}
}

func (r *RequestParam) GetBodyReader() io.Reader {
	if r.Body == nil {
		return nil
	} else {
		return bytes.NewReader(r.Body)
	}
}
func (r *RequestParam) GetBodyStr() string {
	if r.Body == nil {
		return "{}"
	} else {
		return string(r.Body)
	}
}

type Option func(r *RequestParam)

func WithMethod(method string) Option {
	return func(r *RequestParam) {
		r.Method = method
	}
}
func WithContentType(contentType string) Option {
	return func(r *RequestParam) {
		r.ContentType = contentType
	}
}
func WithTimeout(timeout time.Duration) Option {
	return func(r *RequestParam) {
		r.Timeout = timeout
	}
}
func WithBody(body []byte) Option {
	return func(r *RequestParam) {
		r.Body = body
	}
}
func SetHeader(key string, value string) Option {
	return func(r *RequestParam) {
		r.header[key] = value
	}
}

func WithForm() Option {
	return WithContentType(ApplicationForm)
}

func WithQuery(query map[string]string) Option {
	return func(r *RequestParam) {
		r.Query = query
	}
}

func NewRequestParam(ops []Option) *RequestParam {
	req := defaultRequest()
	for _, op := range ops {
		op(req)
	}
	return req
}
