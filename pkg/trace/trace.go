// Package trace provides distributed tracing using OpenTelemetry
// Package trace 提供使用 OpenTelemetry 的分布式追踪
package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer // Global tracer instance | 全局追踪器实例

// Config represents tracing configuration
// Config 表示追踪配置
type Config struct {
	ServiceName string `toml:"service_name"` // Service name | 服务名称
	Endpoint    string `toml:"endpoint"`     // OTLP endpoint, e.g. localhost:4318 | OTLP 端点，例如 localhost:4318
	Insecure    bool   `toml:"insecure"`     // Use insecure connection | 使用不安全连接
}

// Init initializes tracing and returns a shutdown function
// Init 初始化追踪并返回关闭函数
func Init(cfg Config) (func(context.Context) error, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer(cfg.ServiceName)
	return tp.Shutdown, nil
}

// Start creates a child span
// Start 创建子 span
func Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// SpanFromContext gets current span from context
// SpanFromContext 从上下文获取当前 span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span
// AddEvent 向当前 span 添加事件
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
// SetAttributes 在当前 span 上设置属性
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).SetAttributes(attrs...)
}

// TraceID gets current trace ID
// TraceID 获取当前追踪 ID
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
