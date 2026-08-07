package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IResponse describes a response envelope carrying a code and an error.
// Callers unmarshal the JSON response into a type implementing this interface,
// then PlasoPost/PlasoGet verify the code for them.
type IResponse interface {
	GetCode() int
	GetError() error
}

const (
	// CodeOk is the response code treated as success.
	CodeOk = 0
)

// plasoClient is the shared HTTP client used by PlasoPost/PlasoGet.
var plasoClient = &http.Client{Timeout: defaultTimeout}

// PlasoPost marshals reqData to JSON, POSTs it to url, and unmarshals the
// response into res. When res is nil a default envelope is used. It returns
// res.GetError() if the response code is not CodeOk.
func PlasoPost(url string, reqData any, res IResponse, timeout time.Duration) error {
	var data []byte
	if reqData != nil {
		var err error
		if data, err = json.Marshal(reqData); err != nil {
			return err
		}
	}
	if res == nil {
		res = &plasoResponse{}
	}

	body, err := plasoPost(url, data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, res); err != nil {
		return err
	}
	if res.GetCode() != CodeOk {
		return res.GetError()
	}
	return nil
}

// PlasoGet GETs url and unmarshals the response into res. When res is nil a
// default envelope is used. It returns res.GetError() if the response code is
// not CodeOk.
func PlasoGet(url string, res IResponse, timeout time.Duration) error {
	if res == nil {
		res = &plasoResponse{}
	}
	body, err := plasoGet(url)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, res); err != nil {
		return err
	}
	if res.GetCode() != CodeOk {
		return res.GetError()
	}
	return nil
}

func plasoPost(url string, data []byte) ([]byte, error) {
	resp, err := plasoClient.Post(url, ApplicationJSON, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readPlasoBody(resp)
}

func plasoGet(url string) ([]byte, error) {
	resp, err := plasoClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readPlasoBody(resp)
}

func readPlasoBody(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plaso: StatusCode: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// plasoResponse is the default envelope used when the caller passes nil.
type plasoResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r *plasoResponse) GetCode() int {
	return r.Code
}

func (r *plasoResponse) GetError() error {
	if r.Code == CodeOk {
		return nil
	}
	return fmt.Errorf("plaso: code=%d, msg=%s", r.Code, r.Msg)
}
