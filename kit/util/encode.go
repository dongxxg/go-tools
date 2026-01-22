package util

import (
	"net/url"
	"strings"
)

// UrlEncode  对url 进行encode
func UrlEncode(uri string) (string, error) {
	// http://localhost:8989/send?name=wiken&age=18
	index := strings.Index(uri, "?")
	if index < 0 {
		return uri, nil
	}
	u := uri[index+1:]
	lst := strings.Split(u, "&")
	url_ := url.Values{}
	for _, v := range lst {
		temp := strings.Split(v, "=")
		if len(temp) == 2 {
			url_.Add(temp[0], temp[1])
		}
	}
	return uri[:index+1] + url_.Encode(), nil
}
