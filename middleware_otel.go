package sbus

import (
	"context"

	"go.opentelemetry.io/otel"
)

func OtelMiddleware[T input](h HandleFunc[T]) HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		name, ok := ctx.Value("name").(string)

		if !ok {
		}

		ctx, span := otel.Tracer("").Start(ctx, name)
		defer span.End()

		return h(ctx, i)
	}
}
