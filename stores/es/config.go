// 支持ES7接口
package es

import (
	"crypto/tls"
	"net"
	"net/http"

	es "github.com/elastic/go-elasticsearch/v7"
)

type Config struct {
	Addresses []string
	Username  string // 用户名
	Password  string // 密码.
	APIKey    string // 认证用的Base64编码的token
}

// 新建一个不需要认证的配置对象
func NewConfig(addrs []string) *Config {
	return &Config{
		Addresses: addrs,
	}
}

func (c *Config) getESConfig() *es.Config {
	return &es.Config{
		Addresses: c.Addresses,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 0,
			DialContext:           (&net.Dialer{}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS11,
				InsecureSkipVerify: true,
			},
		},
		MaxRetries: 3,
		// EnableMetrics: true,
		RetryOnStatus: []int{502, 503, 504},
		Logger: &Logger{
			EnableRequestBody:  true,
			EnableResponseBody: true,
		},
		Password: c.Password,
		Username: c.Username,
		// Transport:     apmelasticsearch.WrapRoundTripper(http.DefaultTransport),
	}
}
