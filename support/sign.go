package support

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetSign 获取sign
func GetSign(signKey string, params map[string]string, needParam []string) string {
	params["validTime"] = strconv.FormatInt(time.Now().Unix()+30, 10)
	if needParam != nil {
		needParam = append(needParam, "validTime")
	}

	return encrypt(signKey, params, needParam)
}

// GetQueryString 获取querystring
func GetQueryString(params map[string]string, needParam []string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if needParam != nil {
		sort.Strings(needParam)
		keys = needParam
	}

	var res []string
	for _, k := range keys {
		if _, ok := params[k]; !ok {
			continue
		}
		if k == "__plasoRequestId__" {
			continue
		}
		res = append(res, fmt.Sprintf("%s=%s", k, url.QueryEscape(params[k])))
	}

	var content = strings.Join(res[:], "&")
	return content
}

func encrypt(signKey string, params map[string]string, needParam []string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if needParam != nil {
		sort.Strings(needParam)
		keys = needParam
	}

	var res []string
	for _, k := range keys {
		if _, ok := params[k]; !ok || k == "signature" {
			continue
		}
		if k == "__plasoRequestId__" {
			continue
		}
		res = append(res, fmt.Sprintf("%s=%s", k, params[k]))
	}

	var content = strings.Join(res[:], "&")
	var h = hmac.New(sha1.New, []byte(signKey))
	h.Write([]byte(content))
	var signature = strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return signature
}
