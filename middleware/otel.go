package middleware

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
)

func Otel[T any](next sbus.HandleFunc[T]) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		name, ok := ctx.Value("name").(string)

		if !ok {
		}

		ctx, span := otel.Tracer("").Start(ctx, name)
		defer span.End()

		return next(ctx, i)
	}
}

func Otel2[T any](next sbus.HandleFunc[T]) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		name, ok := ctx.Value("name").(string)

		if !ok {
		}

		ctx, span := otel.Tracer("").Start(ctx, name)
		defer span.End()

		return next(ctx, i)
	}
}
