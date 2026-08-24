package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"unitechs.com/unios-dice/uni-base/core/common"
)

const (
	localUserAgent = "UniOS-httpx/1.0.0"

	// Default value for net.Dialer Timeout
	dialTimeout = 30 * time.Second

	// Default value for net.Dialer KeepAlive
	dialKeepAlive = 30 * time.Second

	// Default value for http.Transport TLSHandshakeTimeout
	tlsHandshakeTimeout = 10 * time.Second

	// Default value for Request Timeout
	requestTimeout = 90 * time.Second

	requestRetry = 1

	componentIDGOHttpClient = 5005
)

var pub_otel_enable bool

func init() {
	if os.Getenv("PUB_OTEL_ENABLE") == "true" {
		pub_otel_enable = true
	}
}

type httpclient struct {
	c   *http.Client
	req *http.Request
}

type RequestOptions struct {
	// 如果提供HTTPClient则直接用这个，忽略其他配置。比如skywalking有自己的httpclient
	HTTPClient *http.Client

	// 自定义头
	Headers map[string]string

	// 是否校验服务器的TLS certificate.
	InsecureSkipVerify bool

	// Disable压缩
	DisableCompression bool

	// TLS握手超时
	TLSHandshakeTimeout time.Duration

	// HTTP建立连接超时
	DialTimeout time.Duration

	// KeepAlive 间隔
	DialKeepAlive time.Duration

	// 请求超时
	RequestTimeout time.Duration

	// 配置本端使用的IP地址
	LocalAddr *net.TCPAddr
}

func Request(method, url string, b []byte) (r []byte, err error) {
	ctx := context.Background()
	if b == nil {
		b = []byte{}
	}
	return RequestX(ctx, method, url, b, &RequestOptions{})
}

// RequestX 支持open-tracing
func RequestX(ctx context.Context, method, url string, b []byte, opts *RequestOptions) (r []byte, err error) {
	ctxKey := common.ContextKey(common.HeaderRequestID)
	reqID := ctx.Value(ctxKey)
	if _, ok := reqID.(string); !ok {
		reqID = uuid.New().String()
		ctx = context.WithValue(ctx, ctxKey, reqID)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &RequestOptions{}
	}

	// // skywalking
	// if tracer.GetType() == tracer.TypeSkywalking {
	// 	opts.HTTPClient, err = swhttp.NewClient(tracer.SkyWalkingInstance().GetSkyTracer())
	// 	if err != nil {
	// 		log.Errorf("create skywalking http client error: %s", err)
	// 	}
	// }

	// 将requestid添加到请求头中
	req.Header.Set(common.HeaderRequestID, reqID.(string))
	req.Header.Set("User-Agent", localUserAgent)
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	hc := &httpclient{
		c:   buildHTTPClient(opts),
		req: req,
	}

	// 内部服务设置重试次数
	for i := 0; i <= requestRetry; i++ {
		r, err = hc.request()
		if err == nil {
			break
		}
	}

	return r, nil
}

func (h *httpclient) request() ([]byte, error) {
	if h.c == nil || h.req == nil {
		return nil, errors.New("http client not initialized")
	}

	req := h.req
	var span opentracing.Span

	if opentracing.IsGlobalTracerRegistered() {
		gTracer := opentracing.GlobalTracer()

		// retrieve current Span from Context
		var parentCtx opentracing.SpanContext
		parentSpan := opentracing.SpanFromContext(h.req.Context())
		spanName := "[call] " + req.URL.String()
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
			// start a new Span to wrap HTTP request
			span = gTracer.StartSpan(
				spanName,
				opentracing.ChildOf(parentCtx),
			)
		} else {
			// create a top-level span to represent full work of the client
			span = gTracer.StartSpan(spanName)
			currBin, _ := os.Executable()
			span.SetTag(string(ext.Component), filepath.Base(currBin))
		}

		// make sure the Span is finished once we're done
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)

		req = req.WithContext(ctx)
		var ht *nethttp.Tracer
		// wrap the request in nethttp.TraceRequest
		req, ht = nethttp.TraceRequest(gTracer, req)
		defer ht.Finish()
	}

	if req.Method == "POST" || req.Method == "PUT" {
		req.Header.Set(ContentType, ApplicationJSON)
	}

	res, err := h.c.Do(req)
	if err != nil {
		h.onError(span, err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		h.onError(span, err)
		return body, err
	}
	return body, nil
}

func (h *httpclient) onError(span opentracing.Span, err error) {
	if span != nil {
		// handle errors by recording them in the span
		span.SetTag(string(ext.Error), true)
		span.LogFields(otlog.Error(err))
	}

	// log.Error(err)
}

func RenderWithJSONHeader(w http.ResponseWriter, content []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Write(content)
}

func buildHTTPClient(ro *RequestOptions) *http.Client {
	if ro.HTTPClient != nil {
		return ro.HTTPClient
	}

	// Using the user config for tls timeout or default
	if ro.TLSHandshakeTimeout == 0 {
		ro.TLSHandshakeTimeout = tlsHandshakeTimeout
	}

	// Using the user config for dial timeout or default
	if ro.DialTimeout == 0 {
		ro.DialTimeout = dialTimeout
	}

	// Using the user config for dial keep alive or default
	if ro.DialKeepAlive == 0 {
		ro.DialKeepAlive = dialKeepAlive
	}

	if ro.RequestTimeout == 0 {
		ro.RequestTimeout = requestTimeout
	}

	return &http.Client{
		Transport: CreateHTTPTransport(ro),
		Timeout:   ro.RequestTimeout,
	}
}

// ensureTransporterFinalized will ensure that when the HTTP client is GCed
// the runtime will close the idle connections (so that they won't leak)
// this function was adopted from Hashicorp's go-cleanhttp package
func ensureTransporterFinalized(httpTransport *http.Transport) {
	runtime.SetFinalizer(&httpTransport, func(transportInt **http.Transport) {
		(*transportInt).CloseIdleConnections()
	})
}

type transport struct {
	// tracer    *go2sky.Tracer
	delegated http.RoundTripper
}

func CreateHTTPTransport(opts *RequestOptions) http.RoundTripper {
	dialer := net.Dialer{
		Timeout:   opts.DialTimeout,   // 连接超时时间
		KeepAlive: opts.DialKeepAlive, // 保持长连接的时间
		LocalAddr: opts.LocalAddr,
		DualStack: true,
	}
	defaultTp := &http.Transport{
		DialContext:           dialer.DialContext, // 设置连接的参数
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          500,              // 最大空闲连接
		IdleConnTimeout:       90 * time.Second, // 空闲连接的超时时间
		TLSHandshakeTimeout:   opts.TLSHandshakeTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   10, // 每个host保持的空闲连接数

		// Here comes the user settings
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
		DisableCompression: opts.DisableCompression,
	}
	ensureTransporterFinalized(defaultTp)

	// 不使用skywalking，使用默认transport
	// if tracer.GetType() != tracer.TypeSkywalking {
	// 	return defaultTp
	// }

	// 如果使能skywalking
	ts := transport{
		// tracer:    tracer.SkyWalkingInstance().GetSkyTracer(),
		delegated: defaultTp,
	}
	if pub_otel_enable {
		newts := otelhttp.NewTransport(&ts)
		return newts
	} else {
		return &ts
	}

}

func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// span, err := t.tracer.CreateExitSpan(req.Context(), fmt.Sprintf("/%s%s", req.Method, req.URL.Path),
	// 	req.Host, func(key, value string) error {
	// 		req.Header.Set(key, value)
	// 		return nil
	// 	})
	// if err != nil {
	// 	return t.delegated.RoundTrip(req)
	// }
	// defer span.End()
	// span.SetComponent(componentIDGOHttpClient)
	// for k, v := range t.extraTags {
	// 	span.Tag(go2sky.Tag(k), v)
	// }
	// span.Tag(go2sky.TagHTTPMethod, req.Method)
	// span.Tag(go2sky.TagURL, req.URL.String())
	// span.SetSpanLayer(agentv3.SpanLayer_Http)
	res, err = t.delegated.RoundTrip(req)
	if err != nil {
		// span.Error(time.Now(), err.Error())
		return
	}
	// span.Tag(go2sky.TagStatusCode, strconv.Itoa(res.StatusCode))
	// if res.StatusCode >= http.StatusBadRequest {
	// 	span.Error(time.Now(), "Errors on handling client")
	// }
	return res, nil
}
