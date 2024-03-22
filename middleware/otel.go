package middleware

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
)

type OtelMiddleware[T any] struct{}

func (o *OtelMiddleware[T]) Middleware(h sbus.HandleFunc[T]) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		ctx, span := otel.Tracer("").Start(ctx, "CreateLocationHandler")
		defer span.End()
		// Start span
		// defer span.End()

		return h(ctx, i)
	}
}
