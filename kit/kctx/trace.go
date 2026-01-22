package kctx

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"unitechs.com/cno/kit/component-base/otel"
)

// NewTraceProvider 生成otel provider，
// 在最外层， 要defer cf()
func (c *Ctx) NewTraceProvider(getter IConfigGetter) (cf func(), err error) {
	if err = otel.NewTrace(c, getter); err != nil {
		return
	}
	return func() {
		otel.Shutdown(context.TODO())
	}, nil
}
func (c *Ctx) NewTraceProviderWith(url, serviceName string) (cdf func(), err error) {
	err = otel.NewTraceProvider(url, serviceName)
	return func() {
		otel.Shutdown(context.TODO())
	}, err
}

func (c *Ctx) TraceAndStart(libName, spanName string, opts ...trace.SpanStartOption) trace.Span {
	cc, span := c.TraceAndStartWith(c.Context, libName, spanName, opts...)
	if cc != nil {
		c.Context = cc
	}
	return span
}

func (c *Ctx) TraceAndStartWith(ctx context.Context, libName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.TraceAndStart(ctx, libName, spanName, opts...)
}
func (c *Ctx) SpanIdWith(cc context.Context) string {
	return otel.GetTraceParent(cc)
}
func (c *Ctx) SpanId() string {
	return c.SpanIdWith(c)
}
func (c *Ctx) TraceFromID(spanId, libName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.TraceFromId(spanId, libName, spanName, opts...)
}
func (c *Ctx) TraceFromIdWith(spanId, libName, spanName string, opts ...trace.SpanStartOption) trace.Span {
	cc, span := c.TraceFromID(spanId, libName, spanName, opts...)
	if cc != nil {
		c.Context = cc
	}
	return span

}
