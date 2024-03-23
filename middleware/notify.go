package middleware

import (
	"context"
	"fmt"

	"github.com/loopilot/sbus"
)

func MiddlewareNotify[T any](h sbus.Handler[T], c *sbus.PubsubConfig) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		fmt.Println("notify")
		return h.Next(ctx, i)
	}
}

type Notify struct{}

func WithNotify(n *Notify) sbus.PublishOptionFunc {
	return func(o *sbus.PubsubConfig) {
		o.Metadata().Set("notify", n)

		_, ok := o.Metadata().Get("notify")

		fmt.Println("with notify", ok)
	}
}

// func Otel[T any](next sbus.HandleFunc[T]) sbus.HandleFunc[T] {
// 	return func(ctx context.Context, i T) error {
// 		name, ok := ctx.Value("name").(string)

// 		if !ok {
// 		}

// 		ctx, span := otel.Tracer("").Start(ctx, name)
// 		defer span.End()

// 		return next(ctx, i)
// 	}
// }
