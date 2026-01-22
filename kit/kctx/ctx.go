// 任务号    日期     开发者      描述

package kctx

import (
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"net/http"
	"strings"
	"time"
	"unitechs.com/unios-dice/uni-base/core/log"
)

type Ctx struct {
	log.Logger
	log.Modular
	context.Context
	TraceKey string
	TraceVal string
}

func (c *Ctx) Print(v ...interface{}) {
	c.Info(v...)
}
func (c *Ctx) Printf(format string, v ...interface{}) {
	c.Infof(format, v...)
}
func (c *Ctx) Println(v ...interface{}) {
	c.Info(v...)
}

func (c *Ctx) WithField(key, val string) Ctx {
	c.Modular = c.Modular.With(key, val)
	return *c
}

const TraceId = "trace_id"

// NewCtxWithReq  为request 请求创建ctx
func NewCtxWithReq(ctx context.Context, r *http.Request) Ctx {
	tid := r.Header.Get(TraceId)
	if tid == "" {
		tid = newUUIDStr()
	}

	return NewCtxWithField(ctx, TraceId, tid)
}

// NewCtxWithRequest  为request 请求创建ctx
func NewCtxWithRequest(ctx context.Context, r *http.Request, headerKey string) Ctx {
	tid := r.Header.Get(headerKey)
	if tid == "" {
		tid = newUUIDStr()
	}

	return NewCtxWithField(ctx, headerKey, tid)
}

const spanKey = "Traceparent"

func NewCtxWithOtelRequest(ctx Ctx, r *http.Request, spanName string) (c Ctx, sp trace.Span) {
	spanId := r.Header.Get(spanKey)
	if spanId != "" {
		ctx.Context, sp = ctx.TraceFromID(spanId, "", spanName)
	}
	ctx = NewCtxWithReq(ctx.Context, r)

	return
}

//	func NewCtxWithOtelReq(r *http.Request, spanName string) (c Ctx, sp trace.Span) {
//		c = NewDefaultContext()
//		spanId := r.Header.Get(spanKey)
//		if spanId != "" {
//			c.Context, sp = c.TraceFromID(spanId, "aaaaaa", spanName)
//		} else {
//			sp = otel.NilSpan{}
//		}
//		c = NewCtxWithReq(c.Context, r)
//
//		return
//	}
func newUUIDStr() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

// NewCtxWithRequest2  为request 请求创建ctx
func NewCtxWithRequest2(ctx context.Context, r *http.Request) Ctx {
	return NewCtxWithRequest(ctx, r, TraceId)
}

func NewCtxWithField(ctx context.Context, key, val string) Ctx {
	if val == "" {
		val = newUUIDStr()
	}
	logger := log.GetLogger()

	return Ctx{
		Logger:   *logger,
		TraceKey: key,
		TraceVal: val,
		Modular:  logger.WithFields(map[string]string{key: val}),
		Context:  ctx,
	}
}

func NewContext(ctx context.Context, tid string) Ctx {
	if tid == "" {
		tid = newUUIDStr()
	}
	return NewCtxWithField2(ctx, tid)
}
func NewDefaultContext() Ctx {
	return NewContext(context.TODO(), newUUIDStr())
}

func NewDefaultCtxtWithCancel() (Ctx, context.CancelFunc) {
	c, cf := context.WithCancel(context.TODO())
	return NewContext(c, newUUIDStr()), cf
}

func NewCtxWithField2(ctx context.Context, tid string) Ctx {
	logger := log.GetLogger()
	ctx = context.WithValue(ctx, TraceId, tid)
	return Ctx{
		Logger:   *logger,
		TraceKey: TraceId,
		TraceVal: tid,
		Modular:  logger.WithFields(map[string]string{TraceId: tid}),
		Context:  ctx,
	}
}

func (c *Ctx) GetString(key string) (s string) {
	s, _ = c.Value(key).(string)
	return
}
func (c *Ctx) GetBool(key string) (b bool) {
	b, _ = c.Value(key).(bool)
	return
}
func (c *Ctx) GetInt(key string) (i int) {
	i, _ = c.Value(key).(int)
	return
}
func (c *Ctx) GetInt64(key string) (i int64) {
	i, _ = c.Value(key).(int64)
	return
}
func (c *Ctx) GetUint(key string) (ui uint) {
	ui, _ = c.Value(key).(uint)
	return
}
func (c *Ctx) GetFloat64(key string) (f64 float64) {
	f64, _ = c.Value(key).(float64)
	return
}
func (c *Ctx) GetTime(key string) (t time.Time) {
	t, _ = c.Value(key).(time.Time)
	return
}
func (c *Ctx) GetDuration(key string) (d time.Duration) {
	d, _ = c.Value(key).(time.Duration)
	return
}
func (c *Ctx) GetStringSlice(key string) (ss []string) {
	ss, _ = c.Value(key).([]string)
	return
}
func (c *Ctx) GetStringMap(key string) (sm map[string]interface{}) {
	sm, _ = c.Value(key).(map[string]interface{})
	return
}
func (c *Ctx) GetStringMapString(key string) (sm map[string]string) {
	sm, _ = c.Value(key).(map[string]string)
	return
}
