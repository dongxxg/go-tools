package main

//
//import (
//	"context"
//	"go.opentelemetry.io/otel"
//	"go.opentelemetry.io/otel/attribute"
//	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
//	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
//	"go.opentelemetry.io/otel/propagation"
//	"go.opentelemetry.io/otel/sdk/resource"
//	sdktrace "go.opentelemetry.io/otel/sdk/trace"
//	"go.opentelemetry.io/otel/trace"
//	"time"
//	"unitechs.com/cno/kit/kctx"
//)
//
//func main() {
//
//	if err := New(kctx.NewDefaultContext(), "192.168.5.136:4317", "test"); err != nil {
//		panic(err)
//	}
//	_, span := NewSpan(context.TODO(), "test01")
//	defer span.End()
//	time.Sleep(time.Second * 1)
//}
//
//var tracer = otel.Tracer("unios-alarm3")
//
//func New(ctx kctx.Ctx, OTELEndpoint, serviceName string) error { // todo 改一下服务名
//	exporter, err := otlptrace.New(
//		context.Background(),
//		otlptracegrpc.NewClient(
//			otlptracegrpc.WithInsecure(),
//			otlptracegrpc.WithEndpoint(OTELEndpoint),
//		),
//	)
//	if err != nil {
//		ctx.Fatal(err)
//	}
//	resources, err := resource.New(
//		context.Background(),
//		resource.WithAttributes(
//			attribute.String("service.name", "unios-alarm3"),
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
//
//func NewSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
//	return tracer.Start(ctx, spanName)
//}
//
//func GetTraceParent(c context.Context) string {
//	traceCtx := propagation.TraceContext{}
//	carrier := propagation.MapCarrier{}
//	traceCtx.Inject(c, &carrier)
//	return carrier.Get("traceparent")
//}
//func GetCtxFromTraceParent(traceParent string) context.Context {
//	traceCtx := propagation.TraceContext{}
//	carrier := propagation.MapCarrier{}
//	carrier.Set("traceparent", traceParent)
//	ctx := traceCtx.Extract(context.Background(), &carrier)
//	return ctx // 怎么结束这个span?
//}
//
//// CloseParentSpanFromString 根据父traceparent  关闭总的span
//func CloseParentSpanFromString(traceParent string) {
//	ctx := GetCtxFromTraceParent(traceParent)
//	span := trace.SpanFromContext(ctx)
//	span.End()
//}
