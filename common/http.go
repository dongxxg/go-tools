package common

// 简单包装一下http，用于core包内的调用

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type RequestAction struct {
	Method      string `yaml:"method"`
	URL         string `yaml:"url"`
	Body        string `yaml:"body"`
	Accept      string `yaml:"accept"`
	ContentType string `yaml:"contentType"`
}

func newHTTPRequest(requestAction *RequestAction) (*http.Request, error) {
	if requestAction == nil {
		return nil, fmt.Errorf("empty request, kidding me?")
	}

	var (
		err     error
		request *http.Request
	)

	if requestAction.Body != "" { // add body
		reader := strings.NewReader(requestAction.Body)
		request, err = http.NewRequest(requestAction.Method, requestAction.URL, reader)
	} else {
		request, err = http.NewRequest(requestAction.Method, requestAction.URL, nil)
	}
	if err != nil {
		return nil, err
	}
	// Add headers for request
	request.Header = map[string][]string{
		"Connection": {"keep-alive"},
	}
	if requestAction.Accept != "" {
		request.Header["Accept"] = []string{requestAction.Accept}
	}
	if requestAction.ContentType != "" {
		request.Header["Content-Type"] = []string{requestAction.ContentType}
	}

	return request, nil
}

func DoHTTP(requestAction *RequestAction) (int, []byte, error) {
	request, err := newHTTPRequest(requestAction)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*10) // 设置建立连接超时
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * 10)) // 设置发送接受数据超时
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * 30,
	}

	resp, err := DefaultTransport.RoundTrip(request)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}

	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, bs, err
}
