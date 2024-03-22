package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func getResource() (*resource.Resource, error) {
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("could not merge resources: %w", err)
	}

	return resource, nil
}
