package observe

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const OBSERVCE_MODE_OTEL = "otel"

func Register(applicationName, serviceName, serviceVersion, mode string) error {
	res, err := newOtlpResource(applicationName, serviceName, serviceVersion)
	if err != nil {
		return errors.Join(errors.New("failed to create observability provider: "), err)
	}

	if mode == "otel" {
		tracerProvider, err := newTracerProvider(res)
		if err != nil {
			return errors.Join(errors.New("failed to create observability provider: "), err)
		}
		otel.SetTracerProvider(tracerProvider)
	}

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
	return nil
}

func newOtlpResource(applicationName, serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			attribute.String("env", os.Getenv("ENV")),
			attribute.String("target-application", applicationName),
		))
}

// Not used in this example
func newMetricProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		return nil, errors.Join(errors.New("failed to create metric exporter"), err)
	}
	metricProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	return metricProvider, nil
}

func newTracerProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create trace exporter"), err)
	}
	// Use the exporter
	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(exp),
		trace.WithSampler(DemoSampler{}),
	)
	return traceProvider, nil
}
