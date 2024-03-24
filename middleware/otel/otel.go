package otel

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
)

func Otel[T any](h sbus.Handler, c sbus.PubsubConfig) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		ctx, span := otel.Tracer("opa").Start(ctx, h.Name())
		defer span.End()

		return h.Handle(ctx, i)
	}
}

func WithGroup(g string) sbus.SubscribeOptionFunc {
	return func(o sbus.PubsubConfig) sbus.PubsubConfig {
		o.Metadata().Set("group", g)

		return o
	}
}
