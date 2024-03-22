package sbus

import (
	"context"
	"fmt"
)

// default bus is a global bus that is used by the package
var defaultBus *bus

// input is a type that is used to define the input of a handler
type input interface{}

// bus is a struct that holds the topics and their handlers
type bus struct {
	topics map[string][]interface{}
}

// GetTopic returns the handlers for a given topic
func (b *bus) GetTopic(topic string) ([]interface{}, bool) {
	handlers, ok := b.topics[topic]

	return handlers, ok
}

// addHandler adds a handler to a topic
func (b *bus) addHandler(topic string, h input) error {
	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = make([]interface{}, 0)
	}

	b.topics[topic] = append(b.topics[topic], h)

	return nil
}

// New creates a new bus
func New() *bus {
	b := &bus{
		topics: make(map[string][]interface{}),
	}

	if defaultBus == nil {
		defaultBus = b
	}

	return b
}

// Subscribe adds a handler to a topic
func Subscribe[T input](topic string, fn handlerFunc[T], opts ...subscribeOptionFunc) error {
	c := newSubscribeConfig(opts...)
	h := handler[T]{
		name:        c.name,
		fn:          fn,
		middlewares: c.middlewares,
	}

	if c.bus == nil {
		return ErrBusDoesNotExist
	}

	if err := c.bus.addHandler(topic, h); err != nil {
		return err
	}

	return nil
}

// Publish publishes data to a topic
func Publish[T input](topic string, ctx context.Context, data T) error {
	handlers := defaultBus.topics[topic]

	for _, hndl := range handlers {
		h, ok := hndl.(handler[T])

		if !ok {
			// ErrInvalidHandler
			fmt.Println("INVALID_HANDLER: The handler is not compatible with the topic")
			continue
		}

		if err := h.run(ctx, data); err != nil {
			return err
		}

		// if err := h.fn(ctx, data); err != nil {
		// 	return err
		// }
	}

	return nil
}

// subscribeConfig allows to configure the subscription of a handler
type subscribeConfig struct {
	name        string
	bus         *bus
	middlewares []middlewareFunc
}

type subscribeOptionFunc func(*subscribeConfig)

// WithName sets the name of the handler
func WithName(name string) subscribeOptionFunc {
	return func(c *subscribeConfig) {
		c.name = name
	}
}

// WithBus sets the bus of the handler
func WithBus(b *bus) subscribeOptionFunc {
	return func(c *subscribeConfig) {
		c.bus = b
	}
}

// WithMiddleware adds a middleware to the handler
func WithMiddleware(m middlewareFunc) subscribeOptionFunc {
	return func(c *subscribeConfig) {
		c.middlewares = append(c.middlewares, m)
	}
}

// newSubscribeConfig creates a new subscribeConfig
func newSubscribeConfig(opts ...subscribeOptionFunc) *subscribeConfig {
	c := &subscribeConfig{
		bus: defaultBus,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// handlerFunc is a function that handles the data of a topic
type handlerFunc[T input] func(context.Context, T) error

// handler is a struct that holds the name and the function of a handler
type handler[T input] struct {
	name        string
	fn          handlerFunc[T]
	middlewares []middlewareFunc
}

func (h *handler[T]) run(ctx context.Context, data T) error {
	fmt.Printf("asdasd %T\n", h.fn)

	// middlewareHandler := h.fn

	// middlewareHandler(ctx, data)

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		currentMiddleware := h.middlewares[i]

		_, ok := any(currentMiddleware).(middlewareFunc)

		if !ok {
			fmt.Println("INVALID_MIDDLEWARE: The middleware is not compatible with the handler")
			continue
		}

		// middlewareHandler = xd(middlewareHandler)
	}

	// for _, m := range h.middlewares {
	// 	h.fn = m(h.fn)
	// }

	return h.fn(ctx, data)
}

type middlewareHandlerFunc func(context.Context, input) error

// middlewareFunc is a function that wraps a handlerFunc
type middlewareFunc func(middlewareHandlerFunc) middlewareHandlerFunc
type af[T input] func(handlerFunc[T]) handlerFunc[T]
