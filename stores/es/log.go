package es

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"time"

	"unitechs.com/unios-dice/uni-base/core/log"
)

type Logger struct {
	EnableRequestBody  bool
	EnableResponseBody bool
}

func (l *Logger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	log.Tracef("%s %s %s [status:%d request:%s]\n",
		start.Format(time.RFC3339),
		req.Method,
		req.URL.String(),
		resStatusCode(res),
		dur.Truncate(time.Millisecond),
	)
	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}
		logBodyAsText(&buf, ">")
	}
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)
		logBodyAsText(&buf, "<")
	}
	if err != nil {
		log.Errorf("!ES error: %v\n", err)
	}
	return nil
}

// RequestBodyEnabled returns true when the request body should be logged.
func (l *Logger) RequestBodyEnabled() bool { return l.EnableRequestBody }

// ResponseBodyEnabled returns true when the response body should be logged.
func (l *Logger) ResponseBodyEnabled() bool { return l.EnableResponseBody }

func resStatusCode(res *http.Response) int {
	if res == nil {
		return -1
	}
	return res.StatusCode
}

func logBodyAsText(body io.Reader, prefix string) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		s := scanner.Text()
		if s != "" {
			log.Tracef("%s %s\n", prefix, s)
		}
	}
}
