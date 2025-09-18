package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Meter          metric.Meter
}

func Init(ctx context.Context, serviceName, serviceVersion, environment string) (*Providers, func(context.Context) error, error) {
	// OTel endpoint (OTLP/gRPC) — колектор
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4317" // k8s default (ClusterIP)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("resource.New: %w", err)
	}

	// ---- Traces
	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("otlptracegrpc.New: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp,
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(3*time.Second),
		),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// ---- Metrics
	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("otlpmetricgrpc.New: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(10*time.Second), // dev-friendly
		)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// W3C tracecontext+baggage for any HTTP/RPC
	otel.SetTextMapPropagator(propagation.TraceContext{})

	p := &Providers{
		TracerProvider: tp,
		MeterProvider:  mp,
		Meter:          mp.Meter(serviceName),
	}
	shutdown := func(ctx context.Context) error {
		var e1, e2 error
		e1 = tp.Shutdown(ctx)
		e2 = mp.Shutdown(ctx)
		if e1 != nil {
			return e1
		}
		return e2
	}
	return p, shutdown, nil
}

// Common attributes helper
func CommonAttrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.TelemetrySDKName("opentelemetry"),
	}
}
