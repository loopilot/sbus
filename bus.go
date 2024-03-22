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

// initHandler initializes a topic
func (b *bus) initHandler(topic string) {
	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = make([]interface{}, 0)
	}
}

// addHandler adds a handler to a topic
func (b *bus) addHandler(topic string, h input) error {
	b.initHandler(topic)

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
func Subscribe[T input](topic string, fn handlerFunc[T], opts ...subscribeOption) *handler[T] {
	c := newSubscribeConfig(opts...)

	middlewares := make([]middlewareFunc[T], 0)

	for _, m := range c.middlewares {
		middleware, ok := m.(func(handlerFunc[T]) handlerFunc[T])

		if !ok {
			fmt.Println("INVALID_MIDDLEWARE: The middleware is not compatible with the handler")
			continue
		}

		middlewares = append(middlewares, middleware)
	}

	h := &handler[T]{
		name:        c.name,
		fn:          fn,
		middlewares: middlewares,
	}

	if c.bus == nil {
		c.bus = defaultBus
	}

	if err := c.bus.addHandler(topic, h); err != nil {
		return nil
	}

	return h
}

// Publish publishes data to a topic
func Publish[T input](topic string, ctx context.Context, data T, opts ...publishOption) error {
	c := newPublishConfig(opts...)
	handlers := c.bus.topics[topic]

	for _, hndl := range handlers {
		h, ok := hndl.(*handler[T])

		if !ok {
			// ErrInvalidHandler
			fmt.Println("INVALID_HANDLER: The handler is not compatible with the topic")
			continue
		}

		if err := h.run(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// pubsubConfig allows to configure the subscription and the publishing of a
// handler
type pubsubConfig struct {
	name        string
	bus         *bus
	middlewares []interface{}
}

// subscribeOption applies a configuration to a pubsubConfig. These options are
// only applicable when subscribing to a topic
type subscribeOption interface {
	applySubscribe(*pubsubConfig)
}

type subscribeOptionFunc func(*pubsubConfig)

func (f subscribeOptionFunc) applySubscribe(c *pubsubConfig) {
	f(c)
}

var _ subscribeOption = subscribeOptionFunc(nil)

func newSubscribeConfig(opts ...subscribeOption) *pubsubConfig {
	c := &pubsubConfig{
		bus: defaultBus,
	}

	for _, opt := range opts {
		opt.applySubscribe(c)
	}

	return c
}

// publishOption applies a configuration to a pubsubConfig. These options are
// only applicable when publishing to a topic
type publishOption interface {
	applyPublish(*pubsubConfig)
}

type publishOptionFunc func(*pubsubConfig)

func (f publishOptionFunc) applyPublish(c *pubsubConfig) {
	f(c)
}

var _ publishOption = publishOptionFunc(nil)

func newPublishConfig(opts ...publishOption) *pubsubConfig {
	c := &pubsubConfig{
		bus: defaultBus,
	}

	for _, opt := range opts {
		opt.applyPublish(c)
	}

	return c
}

// pubsubOption applies a configuration to a pubsubConfig. These options are
// applicable when subscribing and publishing to a topic
type pubsubOption interface {
	subscribeOption
	publishOption
}

type pubsubOptionFunc func(*pubsubConfig)

func (f pubsubOptionFunc) applySubscribe(c *pubsubConfig) {
	f(c)
}

func (f pubsubOptionFunc) applyPublish(c *pubsubConfig) {
	f(c)
}

var _ pubsubOption = pubsubOptionFunc(nil)

// WithName sets the name of the handler
func WithName(name string) subscribeOptionFunc {
	return func(c *pubsubConfig) {
		c.name = name
	}
}

// WithBus sets the bus of the handler
func WithBus(b *bus) pubsubOptionFunc {
	return func(c *pubsubConfig) {
		c.bus = b
	}
}

// WithMiddleware adds a middleware to the handler
func WithMiddleware(m interface{}) subscribeOptionFunc {
	return func(c *pubsubConfig) {
		c.middlewares = append(c.middlewares, m)
	}
}

// handlerFunc is a function that handles the data of a topic
type handlerFunc[T input] func(context.Context, T) error

// handler is a struct that holds the name and the function of a handler
type handler[T input] struct {
	name        string
	fn          handlerFunc[T]
	middlewares []middlewareFunc[T]
}

func (h *handler[T]) Use(m ...middlewareFunc[T]) {
	h.middlewares = append(h.middlewares, m...)
}

func (h *handler[T]) run(ctx context.Context, data T) error {
	handler := h.fn

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}

	return handler(ctx, data)
}

// middlewareFunc is a function that wraps a handlerFunc
type middlewareFunc[T input] func(handlerFunc[T]) handlerFunc[T]
