package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer khởi tạo OpenTelemetry tracer provider và đẩy dữ liệu về Jaeger qua gRPC.
// Hàm trả về một function để defer shutdown provider khi ứng dụng đóng.
func InitTracer(serviceName, jaegerEndpoint string) (func(context.Context) error, error) {
	ctx := context.Background()

	// Tạo resource mô tả thông tin service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Tạo exporter đẩy trace về Jaeger bằng giao thức OTLP gRPC
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(jaegerEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Đăng ký TracerProvider
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Thiết lập TracerProvider làm global mặc định
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}
