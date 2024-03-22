package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
)

func SetupTracing(ctx context.Context) (*trace.TracerProvider, error) {
	fmt.Println("Setting up tracing", collectorURL)

	// Set up the exporter
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		return nil, err
	}

	// Set up the resource
	resources, err := getResource()

	if err != nil {
		return nil, err
	}

	// Set the tracer provider
	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resources),
	)

	otel.SetTracerProvider(provider)

	// Return the exporter's shutdown function
	return provider, nil
}
