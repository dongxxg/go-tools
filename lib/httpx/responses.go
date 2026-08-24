package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type CommonError interface {
	Error() string
	SetMsg(msg string) CommonError
	SetTip(tip string) CommonError
	SetData(data any) CommonError
}

func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func OkJSON(w http.ResponseWriter, v interface{}) error {
	return WriteJSON(w, http.StatusOK, v)
}

// WriteJSON 向HTTP响应中写入状态码和JSON body
func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(code)

	if bs, err := json.Marshal(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout has been handled by http.TimeoutHandler,
		// so it's ignored here.
		if err != http.ErrHandlerTimeout {
			return fmt.Errorf("write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		return fmt.Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
	}
	return nil
}

func WriteRawResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                    // 允许访问所有域
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,accept") // header的类型

	w.WriteHeader(status)
	w.Write(data)
}

// UniOS 标准的响应body
type NormalResponse struct {
	// 模块名_功能分组_服务动名词_四位数字编码
	Code string `json:"code,omitempty"`
	// 开发人员友好信息
	// example: success
	Message string `json:"msg,omitempty"`
	// 用户友好提示
	// example: 任务下发成功，10秒倒计时开始
	Tip string `json:"tip,omitempty"`
	// 任意结构的返回数据
	Data interface{} `json:"data,omitempty"`
}

func WriteNormalResponse(w http.ResponseWriter, httpCode int, msg string, v interface{}) error {
	uniCode := "0000"
	if httpCode >= 400 {
		uniCode = fmt.Sprintf("%d", httpCode*100)
	}
	body := NormalResponse{
		Code:    uniCode,
		Message: msg,
		Tip:     msg,
		Data:    v,
	}
	return WriteJSON(w, httpCode, body)
}

func DownloadFileResponse(rw http.ResponseWriter, fileFullPath string) error {
	file, err := os.Open(fileFullPath)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer file.Close()

	fileName := path.Base(fileFullPath)

	// 如果是临时文件，删除临时后缀
	fileName = strings.Split(fileName, TempFilePrefixSeparater)[0]

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))

	_, err = io.Copy(rw, file)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// rw.WriteHeader(http.StatusOK)

	return nil
}

func (s NormalResponse) Error() string {
	return s.Message
}
func (s NormalResponse) SetMsg(msg string) CommonError {
	s.Message = msg
	return s
}
func (s NormalResponse) SetTip(tip string) CommonError {
	s.Tip = tip
	return s
}

func (s NormalResponse) SetData(data any) CommonError {
	s.Data = data
	return s
}
