package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"unitechs.com/unios-dice/uni-base/core/common"
	"unitechs.com/unios-dice/uni-base/lib/mapping"
)

const (
	formKey = "form"
	// pathKey           = "path"
	emptyJSON         = "{}"
	maxMemory         = 32 << 20 // 32MB
	maxBodyLen        = 8 << 20  // 8MB
	separator         = ";"
	tokensInAttribute = 2
)

type contextKey string

var (
	formUnmarshaler = mapping.NewUnmarshaler(formKey, mapping.WithStringValues())
	// pathUnmarshaler = mapping.NewUnmarshaler(pathKey, mapping.WithStringValues())
	pathVars = contextKey("pathVars")
)

// Parses the symbols reside in url path.
// Like http://localhost/bag/{name}
func ParsePath(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Parse the query strings
// Like http://localhost/id?from={from}
func ParseQuery(r *http.Request) map[string][]string {
	return r.URL.Query()
}

// Parses the form request.
func ParseForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}

	params := make(map[string]interface{}, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		}
	}

	return formUnmarshaler.Unmarshal(params, v)
}

func GetHeader(r *http.Request) map[string][]string {
	return r.Header
}

func GetHeaderMeta(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func ParseHeader(headerValue string) map[string]string {
	ret := make(map[string]string)
	fields := strings.Split(headerValue, separator)

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		kv := strings.SplitN(field, "=", tokensInAttribute)
		if len(kv) != tokensInAttribute {
			continue
		}

		ret[kv[0]] = kv[1]
	}

	return ret
}

// Parses the post request which contains json in body.
func ParseJSONBody(r *http.Request, v interface{}) error {
	if withJsonBody(r) {
		reader := io.LimitReader(r.Body, maxBodyLen)
		return mapping.UnmarshalJsonReader(reader, v)
	}

	return mapping.UnmarshalJsonMap(nil, v)
}

func withJsonBody(r *http.Request) bool {
	return r.ContentLength > 0 && strings.Contains(r.Header.Get(ContentType), ApplicationJSON)
}

func Vars(r *http.Request) map[string]string {
	vars, ok := r.Context().Value(pathVars).(map[string]string)
	if ok {
		return vars
	}

	return nil
}

func WithPathVars(r *http.Request, params map[string]string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), pathVars, params))
}

// DecodeIDRequest 通过url中携带id参数来调用底座时，翻译为IDRequest结构
func DecodeIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := ParsePath(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	idInt, err := strconv.ParseUint(id, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("id %s can not be converted to unit64", id)
	}
	return IDRequest{ID: idInt}, nil
}

// ExtractRequestID 从请求头中获取request id
func ExtractRequestID(r *http.Request) string {
	if l, ok := r.Header[common.HeaderRequestID]; ok {
		return l[0]
	}
	return ""
}

// ExtractRequestIDfromCtx 从context中获取request id
func ExtractRequestIDfromCtx(ctx context.Context) string {
	requestID := ""
	ctxKey := common.ContextKey(common.HeaderRequestID)
	reqID := ctx.Value(ctxKey)
	if reqID != nil {
		requestID = reqID.(string)
	}
	return requestID
}

func UploadFileWithTimeout(filename, url string, t time.Duration) (r []byte, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return
	}

	fh, err := os.Open(filename)
	if err != nil {
		return
	}
	defer fh.Close()

	// iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	c := &http.Client{Timeout: t}
	resp, err := c.Post(url, contentType, bodyBuf)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// get response
	return io.ReadAll(resp.Body)
}

// getIP returns the ip address from the http request
func GetClientIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")

	if len(splitIps) > 0 {
		// get last IP in list since ELB prepends other user defined IPs, meaning the last one is the actual client IP.
		netIP := net.ParseIP(splitIps[len(splitIps)-1])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return ip, nil
	}

	return "", errors.New("IP not found")
}
