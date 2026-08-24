package otel

import (
	"context"
	"errors"
	"github.com/dongxxg/go-tools/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"
	"sync"
)

// 紧记，once只会执行一次，无论什么函数，需要执行不同的函数要使用不同的变量
var once sync.Once

func InitOtel(servname string) {
	// 获取opentelemetry的配置
	var otelendpoint string
	// 服务名
	if os.Getenv("PUB_OTEL_ENABLE") == "true" && os.Getenv("PUB_OTEL_ENDPOINT") != "" {
		otelendpoint = os.Getenv("PUB_OTEL_ENDPOINT")
	}
	if otelendpoint != "" && servname != "" {
		go once.Do(func() {
			// open telemetry
			tp, err := initTracer(otelendpoint, servname)
			if err != nil {
				log.Errorf("your first demo error: %s", err.Error())
				return
			}
			defer func() {
				if err := tp.Shutdown(context.Background()); err != nil {
					log.Errorf("Error shutting down tracer provider: %s", err.Error())
				}
			}()
			// log.Infof("otel endpoint: %s servicename: %s", otelendpoint, servname)
			select {}
			// log.Infof("otel2 endpoint: %s servicename: %s", otelendpoint, servname)
		})
	}
}

func initTracer(otelendpoint, servicename string) (*sdktrace.TracerProvider, error) {

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(otelendpoint),
		),
	)
	if err != nil {
		return nil, err
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", servicename),
			// attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	tp := sdktrace.NewTracerProvider(
		// sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		// sdktrace.WithSyncer(exporter),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}