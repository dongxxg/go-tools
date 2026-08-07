package httpx

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type message struct {
	Name string `json:"name"`
}

func TestOk(t *testing.T) {
	w := tracedResponseWriter{
		headers: make(map[string][]string),
	}
	Ok(&w)
	assert.Equal(t, http.StatusOK, w.code)
}

func TestOkJson(t *testing.T) {
	w := tracedResponseWriter{
		headers: make(map[string][]string),
	}
	msg := message{Name: "anyone"}
	OkJSON(&w, msg)
	assert.Equal(t, http.StatusOK, w.code)
	assert.Equal(t, "{\"name\":\"anyone\"}", w.builder.String())
}

func TestWriteJsonTimeout(t *testing.T) {
	// only log it and ignore
	w := tracedResponseWriter{
		headers: make(map[string][]string),
		timeout: true,
	}
	msg := message{Name: "anyone"}
	WriteJSON(&w, http.StatusOK, msg)
	assert.Equal(t, http.StatusOK, w.code)
}

func TestWriteJsonLessWritten(t *testing.T) {
	w := tracedResponseWriter{
		headers:     make(map[string][]string),
		lessWritten: true,
	}
	msg := message{Name: "anyone"}
	WriteJSON(&w, http.StatusOK, msg)
	assert.Equal(t, http.StatusOK, w.code)
}

type tracedResponseWriter struct {
	headers     map[string][]string
	builder     strings.Builder
	code        int
	lessWritten bool
	timeout     bool
}

func (w *tracedResponseWriter) Header() http.Header {
	return w.headers
}

func (w *tracedResponseWriter) Write(bytes []byte) (n int, err error) {
	if w.timeout {
		return 0, http.ErrHandlerTimeout
	}

	n, err = w.builder.Write(bytes)
	if w.lessWritten {
		n -= 1
	}
	return
}

func (w *tracedResponseWriter) WriteHeader(code int) {
	w.code = code
}
