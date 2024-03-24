package otel

import (
	"context"

	"github.com/loopilot/sbus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func Otel[T any](h sbus.Handler, c sbus.PubsubConfig) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		if g, ok := h.Metadata("group"); ok {
			ctx = context.WithValue(ctx, "group", g)
		}

		// when the handler has no group we create a new span linked to the parent span
		if g, ok := h.Metadata("group"); !ok {
			link := trace.LinkFromContext(ctx)
			ctxn := context.Background()
			ctx, span := otel.Tracer("").Start(ctxn, h.Name(), trace.WithLinks(link))
			defer span.End()

			return h.Handle(ctx, i)
		} else {
			// when the handler has a group we check if the group matches the group in the context
			group := g.(string)

			ctxGroup, ok := ctx.Value("group").(string)

			if !ok {
				return nil
			}

			// if the groups match we create a new span linked to the parent span
			if group == ctxGroup {
				ctx, span := otel.Tracer("").Start(ctx, h.Name())
				defer span.End()

				return h.Handle(ctx, i)
			} else {
				link := trace.LinkFromContext(ctx)
				ctxn := context.Background()
				ctx, span := otel.Tracer("").Start(ctxn, h.Name(), trace.WithLinks(link))
				defer span.End()

				return h.Handle(ctx, i)
			}
		}
	}
}

func WithGroup(g string) sbus.SubscribeOptionFunc {
	return func(o sbus.PubsubConfig) sbus.PubsubConfig {
		o.Metadata().Set("group", g)

		return o
	}
}
