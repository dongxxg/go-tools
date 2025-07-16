package http_util

import (
	"encoding/json"
	"errors"
	"strconv"
)

const (
	codeOk = 0
)

type BaseResponse struct {
	Code    int    `json:"code"`
	ReqId   string `json:"reqId"`
	ReqTime int64  `json:"reqTime"`
	ErrMsg  string `json:"errMsg,omitempty"`
}

func (r *BaseResponse) GetCode() int {
	return r.Code
}

func (r *BaseResponse) GetError() error {
	return errors.New("reqId is: " + r.ReqId + ", code is: " + strconv.Itoa(r.Code) + ", errMsg is: " + r.ErrMsg)
}

type iResponse interface {
	GetCode() int
	GetError() error
}

func PlasoPost(url string, reqData interface{}, res iResponse) error {
	var data []byte
	var err error
	if reqData != nil {
		if data, err = json.Marshal(reqData); err != nil {
			return err
		}
	}
	if res == nil {
		res = &BaseResponse{}
	}

	body, err := Post(url, WithBody(data))
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}
	if res.GetCode() != codeOk {
		return res.GetError()
	}
	return nil
}
func PlasoGet(url string, res iResponse, ops ...Option) error {
	if res == nil {
		res = &BaseResponse{}
	}
	body, err := Get(url, ops...)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.GetCode() != codeOk {
		return res.GetError()
	}
	return nil
}
