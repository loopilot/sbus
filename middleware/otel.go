package middleware

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
)

func Otel[T any](h sbus.Handler, c sbus.PubsubConfig) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		ctx, span := otel.Tracer("").Start(ctx, h.Name())
		defer span.End()

		return h.Handle(ctx, i)
	}
}
