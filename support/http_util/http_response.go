package http_util

import (
	"encoding/json"
	"errors"
	"strconv"
)

const (
	CodeOk             = 0
	CodeErrSign        = 1   // 签名不匹配
	CodeErrNoServer    = 10  // 没有可用服务
	CodeErrInner       = 11  // 服务器内部错误
	CodeErrParam       = 12  // 参数错误
	CodeErrNotFound    = 404 // 找不到对象
	CodeMasterNotFound = 101
	CodeMainExecFailed = 102
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

type HttpResponse struct {
	BaseResponse
	Obj any `json:"obj,omitempty"`
}

func (r *HttpResponse) String() string {
	if r == nil {
		return ""
	}
	buf, _ := json.Marshal(r)
	return string(buf)
}

type HttpError struct {
	code   int
	errMsg string
}

func NewHttpError(code int, errMsg string) *HttpError {
	return &HttpError{
		code:   code,
		errMsg: errMsg,
	}
}

func (e *HttpError) GetCode() int {
	return e.code
}

func (e *HttpError) GetErrMsg() string {
	return e.errMsg
}
