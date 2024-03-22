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
	topics map[string][]handler
}

// getTopic returns the handlers for a given topic
func (b *bus) getTopic(topic string) ([]handler, bool) {
	handlers, ok := b.topics[topic]

	return handlers, ok
}

// initHandler initializes a topic
func (b *bus) initHandler(topic string) {
	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = make([]handler, 0)
	}
}

// addHandler adds a handler to a topic
func (b *bus) addHandler(topic string, handle handler) error {
	b.initHandler(topic)

	b.topics[topic] = append(b.topics[topic], handle)

	return nil
}

// New creates a new bus
func New() *bus {
	b := &bus{
		topics: make(map[string][]handler),
	}

	if defaultBus == nil {
		defaultBus = b
	}

	return b
}

// Subscribe adds a handler to a topic
func Subscribe[T input](topic string, fn HandleFunc[T], opts ...subscribeOption) *handle[T] {
	cfg := newSubscribeConfig(opts...)
	hdl := newHandler[T](fn, cfg)

	if cfg.bus == nil {
		if defaultBus == nil {
			New()
		}

		cfg.bus = defaultBus
	}

	if cfg.name == "" {
		cfg.name = fmt.Sprintf("topic: %s", topic)
	}

	if err := cfg.bus.addHandler(topic, hdl); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return nil
	}

	return hdl
}

// Publish publishes data to a topic
func Publish[T input](topic string, ctx context.Context, data T, opts ...publishOption) error {
	cfg := newPublishConfig(opts...)
	bus := cfg.bus
	hdl, ok := bus.getTopic(topic)

	if !ok {
		return ErrTopicDoesNotExist
	}

	for _, hndl := range hdl {
		handle, ok := hndl.(*handle[T])

		if !ok {
			fmt.Println("INVALID_HANDLER: The handler is not compatible with the topic")
			continue
		}

		if err := handle.run(ctx, data); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}
	}

	return nil
}

// pubsubConfig allows to configure the subscription and the publishing of a
// handler
type pubsubConfig struct {
	name string
	bus  *bus
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

// handler is an interface that defines the handler behavior of a handler in
// a bus
type handler interface {
	// run(context.Context, input) error
	HandleFunc() interface{}
	Name() string
}

var _ handler = (*handle[input])(nil)

// handleFunc is a function that handles the data of a topic
type HandleFunc[T input] func(context.Context, T) error

// handle is a struct that holds the name and the function of a handle
type handle[T input] struct {
	name        string
	handleFunc  HandleFunc[T]
	middlewares []middlewareFunc[T]
}

func (h *handle[T]) Use(m ...middlewareFunc[T]) {
	h.middlewares = append(h.middlewares, m...)
}

func (h *handle[T]) run(ctx context.Context, data T) error {
	handleFunc := h.handleFunc

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handleFunc = h.middlewares[i](handleFunc)
	}

	return handleFunc(ctx, data)
}

func (h *handle[T]) Name() string {
	return h.name
}

func (h *handle[T]) HandleFunc() interface{} {
	return h.handleFunc
}

// newHandler creates a new handle
func newHandler[T input](handleFunc HandleFunc[T], cfg *pubsubConfig) *handle[T] {
	return &handle[T]{
		name:        cfg.name,
		handleFunc:  handleFunc,
		middlewares: make([]middlewareFunc[T], 0),
	}
}

// middlewareFunc is a function that wraps a handleFunc
type middlewareFunc[T input] func(HandleFunc[T]) HandleFunc[T]
