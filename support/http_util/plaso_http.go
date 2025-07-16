package http_util

import (
	"encoding/json"
)

type IResponse interface {
	GetCode() int
	GetError() error
}

func PlasoPost(url string, reqData any, res IResponse, ops ...Option) error {
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

	ops = append([]Option{WithBody(data)}, ops...)

	body, err := Post(url, ops...)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, res)
	if err != nil {
		return err
	}
	if res.GetCode() != CodeOk {
		return res.GetError()
	}
	return nil
}
func PlasoGet(url string, res IResponse, ops ...Option) error {
	if res == nil {
		res = &BaseResponse{}
	}
	body, err := Get(url, ops...)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, res); err != nil {
		return err
	}
	if res.GetCode() != CodeOk {
		return res.GetError()
	}
	return nil
}
