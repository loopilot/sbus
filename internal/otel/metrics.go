package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func setupMetrics(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	// Set up the exporter
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(collectorURL),
		otlpmetricgrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	// Set up the resource
	resource, err := getResource()

	if err != nil {
		return nil, err
	}

	// Set up the meter
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)

	return mp, nil
}
