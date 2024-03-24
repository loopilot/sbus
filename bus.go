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
	topics map[string][]Handler
}

// returns the handlers for a given topic
func (b *bus) getHandle(topic string) ([]Handler, bool) {
	handlers, ok := b.topics[topic]

	return handlers, ok
}

// addHandler adds a handler to a topic
func (b *bus) addHandler(topic string, handle Handler) {
	_, ok := b.topics[topic]

	if !ok {
		b.topics[topic] = []Handler{handle}
	} else {
		b.topics[topic] = append(b.topics[topic], handle)
	}
}

// New creates a new bus
func New() *bus {
	b := &bus{
		topics: make(map[string][]Handler),
	}

	if defaultBus == nil {
		defaultBus = b
	}

	return b
}

// Subscribe adds a handler to a topic
func Subscribe[T input](topic string, fn HandleFunc[T], opts ...SubscribeOption) *handle[T] {
	config := NewSubscribeConfig(opts...)
	handle := newHandler[T](fn, config)

	if config.bus == nil {
		if defaultBus == nil {
			New()
		}

		config.bus = defaultBus
	}

	if config.name == "" {
		handle.name = fmt.Sprintf("topic: %s", topic)
	}

	config.bus.addHandler(topic, handle)

	return handle
}

// Publish publishes data to a topic
func Publish[T input](topic string, ctx context.Context, data T, opts ...PublishOption) error {
	config := NewPublishConfig(opts...)
	handle, ok := config.bus.getHandle(topic)

	if !ok {
		return ErrTopicDoesNotExist
	}

	for _, h := range handle {
		if err := h.run(ctx, data, config); err != nil {
			return err
		}
	}

	return nil
}

// PubsubConfig allows to configure the subscription and the publishing of a
// handler
type PubsubConfig struct {
	name     string
	bus      *bus
	metadata metadata
}

// Metadata returns the metadata of the handler
func (c *PubsubConfig) Metadata() metadata {
	return c.metadata
}

// SubscribeOption applies a configuration to a PubsubConfig. These options are
// only applicable when subscribing to a topic
type SubscribeOption interface {
	applySubscribe(PubsubConfig) PubsubConfig
}

type SubscribeOptionFunc func(PubsubConfig) PubsubConfig

func (f SubscribeOptionFunc) applySubscribe(c PubsubConfig) PubsubConfig {
	return f(c)
}

var _ SubscribeOption = SubscribeOptionFunc(nil)

func NewSubscribeConfig(opts ...SubscribeOption) PubsubConfig {
	c := PubsubConfig{
		bus:      defaultBus,
		metadata: make(metadata),
	}

	for _, opt := range opts {
		c = opt.applySubscribe(c)
	}

	return c
}

// PublishOption applies a configuration to a PubsubConfig. These options are
// only applicable when publishing to a topic
type PublishOption interface {
	applyPublish(PubsubConfig) PubsubConfig
}

type PublishOptionFunc func(PubsubConfig) PubsubConfig

func (f PublishOptionFunc) applyPublish(c PubsubConfig) PubsubConfig {
	return f(c)
}

var _ PublishOption = PublishOptionFunc(nil)

func NewPublishConfig(opts ...PublishOption) PubsubConfig {
	c := PubsubConfig{
		bus:      defaultBus,
		metadata: make(metadata),
	}

	for _, opt := range opts {
		c = opt.applyPublish(c)
	}

	return c
}

// PubsubOption applies a configuration to a PubsubConfig. These options are
// applicable when subscribing and publishing to a topic
type PubsubOption interface {
	SubscribeOption
	PublishOption
}

type PubsubOptionFunc func(PubsubConfig) PubsubConfig

func (f PubsubOptionFunc) applySubscribe(c PubsubConfig) PubsubConfig {
	return f(c)
}

func (f PubsubOptionFunc) applyPublish(c PubsubConfig) PubsubConfig {
	return f(c)
}

var _ PubsubOption = PubsubOptionFunc(nil)

// WithName sets the name of the handler
func WithName(name string) SubscribeOptionFunc {
	return func(c PubsubConfig) PubsubConfig {
		c.name = name
		return c
	}
}

// WithBus sets the bus of the handler
func WithBus(b *bus) PubsubOptionFunc {
	return func(c PubsubConfig) PubsubConfig {
		c.bus = b
		return c
	}
}

// WithMetadata sets the metadata of the handler
func WithMetadata(key string, value interface{}) PubsubOptionFunc {
	return func(c PubsubConfig) PubsubConfig {
		c.metadata.Set(key, value)

		return c
	}
}

// metadata is a map that holds the metadata of a pubsub config
type metadata map[string]interface{}

func (m metadata) Get(key string) (interface{}, bool) {
	val, ok := m[key]

	return val, ok
}

func (m metadata) Set(key string, val interface{}) {
	m[key] = val
}

// handler is an interface that defines the handler behavior of a handler
type Handler interface {
	Name() string
	Handle(context.Context, input) error
	Metadata(key string) (interface{}, bool)

	run(context.Context, input, PubsubConfig) error
}

var _ Handler = (*handle[input])(nil)

// handleFunc is a function that handles the data of a topic
type HandleFunc[T input] func(context.Context, T) error

// handle is a struct that holds the name and the function of a handle
type handle[T input] struct {
	name        string
	handleFunc  HandleFunc[T]
	middlewares []middlewareFunc[T]
	metadata    metadata
}

// returns the name of a handle
func (h *handle[T]) Name() string {
	return h.name
}

// executes the handleFunc of a handle
func (h *handle[T]) Handle(ctx context.Context, data input) error {
	t, ok := data.(T)

	if !ok {
		return fmt.Errorf("invalid data type")
	}

	return h.handleFunc(ctx, t)
}

// returns the metadata of a handle
func (h *handle[T]) Metadata(key string) (interface{}, bool) {
	return h.metadata.Get(key)
}

// adds a middleware to a handle
func (h *handle[T]) Use(m ...middlewareFunc[T]) {
	h.middlewares = append(h.middlewares, m...)
}

func (h *handle[T]) run(ctx context.Context, data input, c PubsubConfig) error {
	t, ok := data.(T)

	if !ok {
		return fmt.Errorf("invalid data type")
	}

	handleFunc := h.handleFunc

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handleFunc = h.middlewares[i](h, c)
	}

	return handleFunc(ctx, t)
}

func (h *handle[T]) setupPublishContext(c context.Context) context.Context {
	ctx := context.WithValue(c, "name", h.name)

	return ctx
}

// newHandler creates a new handle
func newHandler[T input](handleFunc HandleFunc[T], cfg PubsubConfig) *handle[T] {
	return &handle[T]{
		name:        cfg.name,
		handleFunc:  handleFunc,
		middlewares: make([]middlewareFunc[T], 0),
		metadata:    cfg.metadata,
	}
}

// middlewareFunc is a function that wraps a handleFunc
type middlewareFunc[T input] func(Handler, PubsubConfig) HandleFunc[T]
