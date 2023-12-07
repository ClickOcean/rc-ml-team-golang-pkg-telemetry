package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type SpanExporter interface {
	ExportSpans(ctx context.Context, spans []tracesdk.ReadOnlySpan) error
	Shutdown(ctx context.Context) error
}

// newExporter creates new OTEL exporter
//
// url example -
// grpc: opentelemetry-collector:4317
// http: opentelemetry-collector:4318
func newExporter(ctx context.Context, url string) (tracesdk.SpanExporter, error) {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(url),
	)
	if err != nil {
		return nil, err
	}
	return traceExporter, nil
}

func newTraceProvider(
	exp tracesdk.SpanExporter,
	ServiceName string) (*tracesdk.TracerProvider, error) {

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	return tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(r),
	), nil
}

func Init(ctx context.Context, exporterURL string, serviceName string) (trace.Tracer, error) {
	exporter, err := newExporter(ctx, exporterURL)
	if err != nil {
		return nil, err
	}

	tp, err := newTraceProvider(exporter, serviceName)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp) // !!!!!!!!!!!
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer := tp.Tracer(serviceName)

	// Start a goroutine to listen for context cancellation and shutdown the tracer provider
	go func() {
		<-ctx.Done()
		if err := tp.Shutdown(ctx); err != nil && err != context.Canceled {
			panic(err)
		}
	}()

	return tracer, nil
}
