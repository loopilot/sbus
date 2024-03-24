package middleware

import (
	"context"
	"fmt"

	"github.com/loopilot/sbus"
)

// func lowerMiddleware[T string](h Handler, c PubsubConfig) HandleFunc[T] {
// 	return func(ctx context.Context, data T) error {
// 		str, ok := any(data).(string)

// 		if !ok {
// 			return fmt.Errorf("expected data to be of type string")
// 		}

// 		lower := strings.ToLower(str)

// 		return h.Next(ctx, T(lower))

// 		// return next(ctx, T(lower))
// 	}
// }

func MiddlewareNotify[T any](h sbus.Handler, c sbus.PubsubConfig) sbus.HandleFunc[T] {
	return func(ctx context.Context, i T) error {
		fmt.Println("notify")
		return h.Handle(ctx, i)
	}
}

type Notify struct{}

func WithNotify(n *Notify) sbus.PublishOptionFunc {
	return func(o sbus.PubsubConfig) sbus.PubsubConfig {
		o.Metadata().Set("notify", n)

		_, ok := o.Metadata().Get("notify")

		fmt.Println("with notify", ok)

		return o
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
