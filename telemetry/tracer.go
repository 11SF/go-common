package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	ExporterType   string // "stdout", "jaeger", "zipkin"
}

type Telemetry struct {
	tracer         oteltrace.Tracer
	tracerProvider *trace.TracerProvider
	config         Config
}

func New(config Config) (*Telemetry, error) {
	if config.ServiceName == "" {
		config.ServiceName = "unknown-service"
	}
	if config.ServiceVersion == "" {
		config.ServiceVersion = "unknown"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	if config.ExporterType == "" {
		config.ExporterType = "stdout"
	}

	exporter, err := createExporter(config.ExporterType)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(config.ServiceName)

	return &Telemetry{
		tracer:         tracer,
		tracerProvider: tp,
		config:         config,
	}, nil
}

func createExporter(exporterType string) (trace.SpanExporter, error) {
	switch exporterType {
	case "stdout":
		return stdouttrace.New(
			stdouttrace.WithWriter(os.Stdout),
		)
	default:
		return stdouttrace.New(
			stdouttrace.WithWriter(os.Stdout),
		)
	}
}

func (t *Telemetry) Tracer() oteltrace.Tracer {
	return t.tracer
}

func (t *Telemetry) StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *Telemetry) SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func (t *Telemetry) TraceID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

func (t *Telemetry) SpanID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}

func (t *Telemetry) SetSpanAttributes(ctx context.Context, attrs ...oteltrace.SpanStartOption) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	// Note: SpanStartOption can't be applied after span creation
	// This method is for future use with attribute options
}

func (t *Telemetry) AddEvent(ctx context.Context, name string, attrs ...oteltrace.EventOption) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.AddEvent(name, attrs...)
}

func (t *Telemetry) RecordError(ctx context.Context, err error, attrs ...oteltrace.EventOption) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.RecordError(err, attrs...)
}

func (t *Telemetry) SetStatus(ctx context.Context, code codes.Code, description string) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.SetStatus(code, description)
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.tracerProvider.Shutdown(ctx)
}

var globalTelemetry *Telemetry

func Init(config Config) error {
	tel, err := New(config)
	if err != nil {
		return err
	}
	globalTelemetry = tel
	return nil
}

func Global() *Telemetry {
	return globalTelemetry
}

func StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	if globalTelemetry == nil {
		return ctx, oteltrace.SpanFromContext(ctx)
	}
	return globalTelemetry.StartSpan(ctx, name, opts...)
}

func TraceID(ctx context.Context) string {
	if globalTelemetry == nil {
		return ""
	}
	return globalTelemetry.TraceID(ctx)
}

func SpanID(ctx context.Context) string {
	if globalTelemetry == nil {
		return ""
	}
	return globalTelemetry.SpanID(ctx)
}

func RecordError(ctx context.Context, err error, attrs ...oteltrace.EventOption) {
	if globalTelemetry == nil {
		return
	}
	globalTelemetry.RecordError(ctx, err, attrs...)
}

func AddEvent(ctx context.Context, name string, attrs ...oteltrace.EventOption) {
	if globalTelemetry == nil {
		return
	}
	globalTelemetry.AddEvent(ctx, name, attrs...)
}
