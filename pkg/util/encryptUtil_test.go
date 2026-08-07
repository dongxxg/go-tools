package util

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestEncrypt(t *testing.T) {
	params := make(map[string]string)
	params["validTime"] = "3600"
	params["validBegin"] = "1632837592"
	params["appId"] = "plaso"
	params["fileId"] = "6058407de2d1c42f8452d630"
	appkey := "dt47593e115f8f1b5a185216400f31261_wyy"
	fmt.Println(Encrypt(params, appkey))
}

func TestEncrypt2(t *testing.T) {
	params := make(map[string]string)
	validBegin := time.Now().Unix()
	fmt.Println(validBegin)
	params["validTime"] = "3600"
	params["validBegin"] = strconv.FormatInt(validBegin, 10)
	params["appId"] = "plaso"
	params["remoteDir"] = "dev-plaso/liveclass/1866/wyy_1632453689038_dev(Go)"
	appkey := "dt47593e115f8f1b5a185216400f31261_wyy"
	fmt.Println(Encrypt(params, appkey))
}

func TestSignQuery(t *testing.T) {
	// appKey和服务器保持一致
	var query = "appId=plaso&endTime=1696215153000&groupId=0&" +
		"loginName=wyy&meetingId=wyymid&meetingType=5&userName=%E6%A2%81%E5%85%86%E4%BA%AE&userType=webVisitor"
	queryValue, _ := url.ParseQuery(query)
	queryValue.Set("validBegin", strconv.FormatInt(time.Now().UnixMilli(), 10))
	queryValue.Set("validTime", strconv.FormatInt(10*60*1000, 10))
	signature, _ := EncryptUrlValue(queryValue, "d44")
	queryValue.Set("signature", signature)
	queryRes := queryValue.Encode()
	var host = "http://dev-s1.plaso.cn/school/balance/getMeetingConfigBySign?"
	fmt.Println(host + queryRes)
}
