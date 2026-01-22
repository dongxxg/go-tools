package otel

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"go.opentelemetry.io/otel/trace/noop"
)

type IConfigGetter interface {
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
}

type ILogger interface {
	Infof(v string, args ...interface{})
}

type NilTrace struct {
	embedded.Tracer
}

var noopTracer = noop.NewTracerProvider().Tracer("")

func (n NilTrace) tracer() {
}

func (n NilTrace) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return noopTracer.Start(ctx, spanName, opts...)
}

type NilSpan struct {
	embedded.Span
}

func (n NilSpan) span() {
}

func (n NilSpan) End(options ...trace.SpanEndOption) {
}

func (n NilSpan) AddEvent(name string, options ...trace.EventOption) {
}

func (n NilSpan) IsRecording() bool {
	return false
}

func (n NilSpan) RecordError(err error, options ...trace.EventOption) {
}

func (n NilSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

func (n NilSpan) SetStatus(code codes.Code, description string) {
}

func (n NilSpan) SetName(name string) {
}

func (n NilSpan) SetAttributes(kv ...attribute.KeyValue) {
}

func (n NilSpan) TracerProvider() trace.TracerProvider {
	return NilTraceProvider{}
}

type NilTraceProvider struct {
	embedded.TracerProvider
}

func (n NilTraceProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return NilTrace{}
}
