package main

import (
	"context"
	"errors"
	"flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
)

const (
	service     = "trace-demo"
	environment = "production"
	id          = 1
)

func main() {
	url := flag.String("jaeger", "http://127.0.0.1:14268/api/traces", "")
	tp, err := traceProvider(*url)
	if err != nil {
		log.Fatal(err)
	}
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ctx context.Context) {
		if err = tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	otel.SetTracerProvider(tp)

	// 随context传数据
	m0, _ := baggage.NewMember("data1", "value1")
	m1, _ := baggage.NewMember("data2", "value2")
	b, _ := baggage.New(m1, m0)
	ctx = baggage.ContextWithBaggage(ctx, b)

	tr := otel.Tracer("com-main")
	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	err = bar(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func traceProvider(url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		log.Fatal(err)
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			attribute.String("environment", environment), // 自定义属性
			attribute.Int64("ID", id),
		),
		),
	)
	return tp, nil
}

func bar(ctx context.Context) error {
	tr := otel.Tracer("com-bar")
	_, span := tr.Start(ctx, "bar")
	defer span.End()

	// 业务逻辑
	span.SetAttributes(attribute.Key("test").String("value"))
	span.SetAttributes(attribute.Key(baggage.FromContext(ctx).Member("data1").Key()).String(baggage.FromContext(ctx).Member("data1").Value()))

	// 测试error
	err := errors.New("bar error")
	span.AddEvent(err.Error())
	span.SetStatus(codes.Error, err.Error())
	return err
}
