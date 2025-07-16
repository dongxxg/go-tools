package http_util

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	url := "https://dev.plaso.cn/bbs/getBoardConfig?appId=1866&signature=A37083707DC5EB21CFBC25160D5997D1D60D5FEB&ownerLoginName=hqmin1&recordId=a123&loginName=hqmin&userType=editor"
	res, _ := Get(url)
	fmt.Println("res", string(res))
}

func TestDown2File(t *testing.T) {
	url := "https://file-plaso.oss-cn-hangzhou.aliyuncs.com/dev-plaso/infinite_wb/base/1866/hqmin/a1234/meta.pb"
	localFile := "./meta.pb"
	err := Down2File(url, localFile)
	fmt.Printf("error: %v\n", err)
}

func TestDown2Memory(t *testing.T) {
	url := "https://file-plaso.oss-cn-hangzhou.aliyuncs.com/dev-plaso/infinite_wb/pdf/A0.pdf"
	content, suffix, err := Down2Memory(url)
	if len(content) > 10 {
		content = content[:10]
	}
	fmt.Println("content:", string(content))
	fmt.Println("suffix:", string(suffix))
	fmt.Printf("error: %v\n", err)
}

func TestTimeout(t *testing.T) {
	url := "https://file-plaso.oss-cn-hangzhou.aliyuncs.com/dev-plaso/infinite_wb/base/1866/hqmin/a1234/meta.pb"
	localFile := "./meta.pb"
	_ = Down2File(url, localFile, WithTimeout(time.Millisecond))
}

func TestWithBody(t *testing.T) {
	url := "https://dev.plaso.cn/bbs/getBoardConfig?appId=1866&signature=A37083707DC5EB21CFBC25160D5997D1D60D5FEB&ownerLoginName=hqmin1&recordId=a123&loginName=hqmin&userType=editor"
	user := map[string]interface{}{"name": "bb"}
	body, _ := json.Marshal(user)
	res, _ := Get(url, WithBody(body))
	fmt.Println("res", string(res))
}

func TestAddQuery(t *testing.T) {
	url := "https://dev.plaso.cn/bbs/getBoardConfig"
	url2 := addQuery(url, map[string]string{"name": "jack"})
	fmt.Println(url2)

	url = "https://dev.plaso.cn/bbs/getBoardConfig?appId=1866"
	url2 = addQuery(url, map[string]string{"name": "jack"})
	fmt.Println(url2)
}
