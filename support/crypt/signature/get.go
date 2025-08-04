package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"
)

// BaseHttpClient 实现了 HttpClientInterface，并提供加密功能
type BaseHttpClient struct {
	AppId   string //应用id
	SignKey string //签名key
}

func GenerateQuery(needEncode bool, params map[string]string) string {
	businessParam := make(map[string]string)
	/*遍历params,将params中的kV值放入businessParam,并将value转为string*/
	for key, value := range params {
		businessParam[key] = fmt.Sprintf("%v", value)
	}
	// 业务参数和签名验证混合后排序
	var keys []string
	for key := range businessParam {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var res []string
	for _, key := range keys {
		if _, ok := businessParam[key]; !ok {
			continue
		}
		if key == "__plasoRequestId__" {
			continue
		}
		if needEncode {
			res = append(res, fmt.Sprintf("%s=%s", key, url.QueryEscape(businessParam[key])))
		} else {
			res = append(res, fmt.Sprintf("%s=%s", key, businessParam[key]))
		}
	}
	// 排序后连接成字符串
	content := ""
	if len(res) > 0 {
		content = res[0]
		for i := 1; i < len(res); i++ {
			content += "&" + res[i]
		}
	}
	return content
}

// 只对参数进行签名

func (_this *BaseHttpClient) Encrypt(params map[string]string) (string, string, error) {
	if len(_this.SignKey) == 0 {
		return "", "", fmt.Errorf("miss signKey")
	}
	// 参与签名的内容不需要encode，
	signContent := GenerateQuery(false, params)
	h := hmac.New(sha1.New, []byte(_this.SignKey))
	h.Write([]byte(signContent))
	var signature = fmt.Sprintf("%X", h.Sum(nil))
	return signature, GenerateQuery(
		true,
		params,
	) + "&signature=" + signature, nil
}

// 检查签名

func (_this *BaseHttpClient) CheckSign(params map[string]string) error {
	sign := params["signature"]
	delete(params, "signature")
	signature, _, err := _this.Encrypt(params)
	if err != nil {
		return err
	}
	log.Println(signature)
	if sign != signature {
		return errors.New("sign not match")
	}
	return nil
}

func main() {
	client := &BaseHttpClient{
		AppId:   "rk1",
		SignKey: "dt47593e1156711b5a185216400f3126",
	}
	params := make(map[string]string)
	params["appId"] = "rk1"
	expire := fmt.Sprintf("%d", time.Now().UnixMicro()+60000)
	params["expire"] = expire
	params["creatorId"] = "655c7eb311624b0001ebfd06"
	fmt.Println(params)
	client.CheckSign(params)
}
