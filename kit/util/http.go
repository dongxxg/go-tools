package util

import (
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
	"unitechs.com/cno/kit/kctx"
)

var myClient *http.Client

func HttpDo(req *http.Request) (*http.Response, error) {
	if myClient == nil {
		myClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}
	return myClient.Do(req)
}

func HttpDoWithResByte(ctx kctx.Ctx, req *http.Request) ([]byte, error) {
	if myClient == nil {
		myClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	res, err := myClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "http do failed")
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read all failed")
	}
	return b, nil
}
