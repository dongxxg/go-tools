package otel

import (
	"context"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

var tp *sdktrace.TracerProvider

func NewTrace(log ILogger, getter IConfigGetter) error {
	if getter.GetBool("Pub.Otel.Enabled") {
		uri := getter.GetString("Pub.Otel.Endpoint")
		if uri == "" {
			return errors.New("Pub.Otel.Endpoint is empty")
		}
		SrvName := getter.GetString("SD.Name")
		if SrvName == "" {
			SrvName = "unios-alarm"
		}
		log.Infof("init otel trace %p %p", uri, SrvName)
		return NewTraceProvider(uri, SrvName)
	}
	return nil
}

func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if tp != nil {
		return tp.Tracer(name, opts...)
	}
	return NilTrace{}
}

func TraceAndStart(ctx context.Context, libName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tp != nil {
		tr := tp.Tracer(libName)
		return tr.Start(ctx, spanName, opts...)
	}
	return noopTracer.Start(ctx, spanName, opts...)
}
func TraceFromId(spanId, libName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx := GetCtxFromTraceParent(spanId)
	return TraceAndStart(ctx, libName, spanName, opts...)
}

var once = sync.Once{}

// NewTraceProvider
// url http://192.168.5.136:14268/api/traces
// serviceName: 根服务名称
func NewTraceProvider(url, serviceName string) (err error) {
	once.Do(func() {
		var p *sdktrace.TracerProvider
		p, err = NewProvider(url, serviceName)
		if err != nil {
			return
		}
		tp = p
		otel.SetTracerProvider(tp)
	})
	return
}
func Provider() (*sdktrace.TracerProvider, bool) {
	if tp == nil {
		return tp, false
	}
	return tp, true
}

func Shutdown(ctx context.Context) error {
	if tp != nil {
		return tp.Shutdown(ctx)
	}
	return nil
}

func NewProvider(url string, rootName string) (*sdktrace.TracerProvider, error) {
	return tracerProvider(url, rootName)
}
func tracerProvider(url string, rootName string) (*sdktrace.TracerProvider, error) {
	// 创建jaeger provider
	// 可以直接连collector也可以连agent
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	p := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(rootName),
			attribute.String("environment", "test"),
			//attribute.Int64("ID", 1),
		)),
	)
	return p, err
}

// ++++++++++++++++++++++++++++++++++++
//
//func New(ctx kctx.Ctx, OTELEndpoint, serviceName string) error { // todo 改一下服务名
//	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(OTELEndpoint)))
//	//exporter, err := otlptrace.New(
//	//	context.Background(),
//	//	otlptracegrpc.NewClient(
//	//		otlptracegrpc.WithInsecure(),
//	//		otlptracegrpc.WithEndpoint(OTELEndpoint),
//	//	),
//	//)
//
//	if err != nil {
//		ctx.Fatal(err)
//	}
//	resources, err := resource.New(
//		context.Background(),
//		resource.WithAttributes(
//			attribute.String("service.name", serviceName),
//		),
//	)
//	if err != nil {
//		ctx.Errorf("Could not set resources: %s", err.Error())
//	}
//
//	tp := sdktrace.NewTracerProvider(
//		// sdktrace.WithSampler(sdktrace.AlwaysSample()),
//		// sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
//		// sdktrace.WithSyncer(exporter),
//		sdktrace.WithBatcher(exporter),
//		sdktrace.WithResource(resources),
//	)
//
//	otel.SetTracerProvider(tp)
//	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
//	return nil
//}

func GetTraceParent(c context.Context) string {
	traceCtx := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	traceCtx.Inject(c, &carrier)
	return carrier.Get("traceparent")
}
func GetCtxFromTraceParent(traceParent string) context.Context {
	traceCtx := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	carrier.Set("traceparent", traceParent)
	ctx := traceCtx.Extract(context.Background(), &carrier)
	return ctx // 怎么结束这个span?
}

// CloseParentSpanFromString 根据父traceparent  关闭总的span
func CloseParentSpanFromString(traceParent string) {
	ctx := GetCtxFromTraceParent(traceParent)
	span := trace.SpanFromContext(ctx)
	span.End()
}
