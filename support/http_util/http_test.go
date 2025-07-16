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

// pptProxy下载失败，不会有err
// 通过文件小于1KB判断下载失败
func TestPptProxy(t *testing.T) {
	url := "https://www.plaso.cn/ppt/getPdf?furl=https%3A%2F%2Fboard-infi.infi.cn%2Finfinite_wb%2Ffiles%2Finfi%2F6247bf70bf6bda0995a09601_8_1_7_1688099334652.xlsx"
	localFile := "./a.pdf"
	err := Down2File(url, localFile)
	fmt.Println("err:", err)
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
	user := map[string]any{"name": "bb"}
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

func TestDownloadJsonError(t *testing.T) {
	// url := "https://dev.plaso.cn/ppt/getPdf?furl=https%3A%2F%2Ffile.plaso.cn%2Fdev-plaso%2Finfinite_wb%2Fpdf%2Fdownload.PDF1"
	url := "https://dev.plaso.cn/ppt/getPdf?furl=abc"
	body, ext, err := Down2Memory(url)
	fmt.Println("body:", string(body))
	fmt.Println("ext:", ext)
	fmt.Println("err:", err)
}
func TestDownloadError(t *testing.T) {
	url := "https://file-plaso.oss-cn-hangzhou.aliyuncs.com/dev-plaso/infinite_wb/files/infi/62440544a797e02d61049f68_8_1_157_1677206464876.pdf_i/1.jpg?x-oss-process=image/resize,h_90"
	Down2Memory(url)
}
