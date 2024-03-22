package middleware

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
)

type OtelMiddleware[T any] struct{}

func NewOtelMiddleware[T any]() *OtelMiddleware[T] {
	return &OtelMiddleware[T]{}
}

func (o *OtelMiddleware[T]) Middleware(h sbus.HandleFunc[T]) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		name, ok := ctx.Value("name").(string)

		if !ok {
		}

		ctx, span := otel.Tracer("").Start(ctx, name)
		defer span.End()

		return h(ctx, i)
	}
}
