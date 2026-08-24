package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"unitechs.com/unios-dice/uni-base/core/log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ValidTimeSignParam = "validTime"
const ValidBeginSignParam = "validBegin"
const SignatureSignParam = "signature"

var signParamName = []string{ValidTimeSignParam, ValidBeginSignParam, SignatureSignParam}

func EncryptUrlValue(urlparams url.Values, appkey string) (string, error) {
	params := make(map[string]string)
	for k, v := range urlparams {
		params[k] = v[0]
	}
	return Encrypt(params, appkey)
}

func Encrypt(params map[string]string, appkey string) (string, error) {
	if len(appkey) == 0 {
		return "", errors.New("miss appkey")
	}
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	keyValues := []string{}
	for _, k := range keys {
		if k == "signature" || len(k) == 0 {
			continue
		}
		keyValues = append(keyValues, k+"="+url.QueryEscape(params[k]))
	}
	var p = strings.Join(keyValues, "&")
	log.Debug("encrypt string is %s", p)
	mac := hmac.New(sha1.New, []byte(appkey))
	mac.Write([]byte(p))
	var signature = fmt.Sprintf("%X", mac.Sum(nil))
	return signature, nil
}

func CheckValidQuery(query url.Values, getAppKey func() string) bool {
	ok, _ := CheckValidQueryWithErrMsg(query, getAppKey)
	return ok
}

func CheckValidQueryWithErrMsg(query url.Values, getAppKey func() string) (bool, error) {
	//签名有效期
	validTime, _ := strconv.ParseInt(query.Get("validTime"), 10, 64)
	//签名开始时间
	validBegin, _ := strconv.ParseInt(query.Get("validBegin"), 10, 64)
	if (validBegin + validTime) < time.Now().UnixMilli() {
		log.Error("check token %s valid time exceeded", query.Encode())
		return false, errors.New("signature valid time exceeded")
	}
	sign := query.Get("signature")
	if tmpSign, err := EncryptUrlValue(query, getAppKey()); err == nil {
		if tmpSign == sign {
			log.Debug("check token ok, %s", query)
			return true, nil
		} else {
			log.Errorf("check token failed,%s invalid signature expect %s, appkey is %s", query, tmpSign, EncryptKey(getAppKey()))
			return false, errors.New("sign is not match")
		}
	} else {
		log.Errorf("check token failed,%s as encrypt faile %s", query, err)
		return false, errors.New("check sign failed")
	}
}

func checkSignParam(query map[string]string, signParamName []string) bool {
	for _, pName := range signParamName {
		if len(query[pName]) == 0 {
			log.Warn("miss signparam %s", pName)
			return false
		}
	}
	return true
}

func CheckValidParam(query map[string]string, getAppKey func() string) bool {
	if !checkSignParam(query, signParamName) {
		return false
	}
	//签名有效期
	validTime, _ := strconv.ParseInt(query[ValidTimeSignParam], 10, 64)
	//签名开始时间
	validBegin, _ := strconv.ParseInt(query[ValidBeginSignParam], 10, 64)
	if (validBegin + validTime) < time.Now().UnixMilli() {
		queryByte, _ := json.Marshal(query)
		log.Error("check token valid time exceeded ," + string(queryByte))
		return false
	}
	sign := query[SignatureSignParam]
	if tmpSign, err := Encrypt(query, getAppKey()); err == nil {
		if tmpSign == sign {
			log.Debug("check token ok, %s ", query)
			return true
		} else {
			queryByte, _ := json.Marshal(query)
			log.Errorf("check token failed, invalid signature expect %s %s", string(queryByte), tmpSign)
			return false
		}
	} else {
		queryByte, _ := json.Marshal(query)
		log.Errorf("check token failed,as encrypt failed %s %s ", string(queryByte), err)
		return false
	}
}
