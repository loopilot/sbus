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

// returns the handlers for a given topic
func (b *bus) GetTopic(topic string) ([]handler, bool) {
	handlers, ok := b.topics[topic]

	return handlers, ok
}

// initializes a topic
func (b *bus) initHandler(topic string) {
	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = make([]handler, 0)
	}
}

// addHandler adds a handler to a topic
func (b *bus) addHandler(topic string, handle handler) {
	b.initHandler(topic)

	b.topics[topic] = append(b.topics[topic], handle)
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
func Subscribe[T input](topic string, fn HandleFunc[T], opts ...SubscribeOption) *handle[T] {
	cfg := NewSubscribeConfig(opts...)
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

	cfg.bus.addHandler(topic, hdl)

	return hdl
}

// Publish publishes data to a topic
func Publish[T input](topic string, ctx context.Context, data T, opts ...PublishOption) error {
	cfg := NewPublishConfig(opts...)
	bus := cfg.bus
	hdl, ok := bus.GetTopic(topic)

	if !ok {
		return ErrTopicDoesNotExist
	}

	for _, hndl := range hdl {
		handle, ok := hndl.(*handle[T])

		if !ok {
			fmt.Println("INVALID_HANDLER: The handler is not compatible with the topic")
			continue
		}

		if err := handle.run(ctx, data, cfg); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
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

func (c *PubsubConfig) Metadata() metadata {
	return c.metadata
}

// SubscribeOption applies a configuration to a PubsubConfig. These options are
// only applicable when subscribing to a topic
type SubscribeOption interface {
	applySubscribe(*PubsubConfig)
}

type SubscribeOptionFunc func(*PubsubConfig)

func (f SubscribeOptionFunc) applySubscribe(c *PubsubConfig) {
	f(c)
}

var _ SubscribeOption = SubscribeOptionFunc(nil)

func NewSubscribeConfig(opts ...SubscribeOption) *PubsubConfig {
	c := &PubsubConfig{
		bus:      defaultBus,
		metadata: make(metadata),
	}

	for _, opt := range opts {
		opt.applySubscribe(c)
	}

	return c
}

// PublishOption applies a configuration to a PubsubConfig. These options are
// only applicable when publishing to a topic
type PublishOption interface {
	applyPublish(*PubsubConfig)
}

type PublishOptionFunc func(*PubsubConfig)

func (f PublishOptionFunc) applyPublish(c *PubsubConfig) {
	f(c)
}

var _ PublishOption = PublishOptionFunc(nil)

func NewPublishConfig(opts ...PublishOption) *PubsubConfig {
	c := &PubsubConfig{
		bus:      defaultBus,
		metadata: make(metadata),
	}

	for _, opt := range opts {
		opt.applyPublish(c)
	}

	return c
}

// PubsubOption applies a configuration to a PubsubConfig. These options are
// applicable when subscribing and publishing to a topic
type PubsubOption interface {
	SubscribeOption
	PublishOption
}

type PubsubOptionFunc func(*PubsubConfig)

func (f PubsubOptionFunc) applySubscribe(c *PubsubConfig) {
	f(c)
}

func (f PubsubOptionFunc) applyPublish(c *PubsubConfig) {
	f(c)
}

var _ PubsubOption = PubsubOptionFunc(nil)

// WithName sets the name of the handler
func WithName(name string) SubscribeOptionFunc {
	return func(c *PubsubConfig) {
		c.name = name
	}
}

// WithBus sets the bus of the handler
func WithBus(b *bus) PubsubOptionFunc {
	return func(c *PubsubConfig) {
		c.bus = b
	}
}

// WithMetadata sets the metadata of the handler
func WithMetadata(key string, value interface{}) PubsubOptionFunc {
	return func(c *PubsubConfig) {
		c.metadata.Set(key, value)
	}
}

// handler is an interface that defines the handler behavior of a handler in
// a bus
type handler interface {
	Name() string
	Metadata(key string) (interface{}, bool)

	// run(context.Context, input) error
	// HandleFunc() interface{}
	// HandleFunc() interface{}
}

type Handler[T input] interface {
	Next(ctx context.Context, data T) error
	Name() string
}

var _ handler = (*handle[input])(nil)
var _ Handler[input] = (*handle[input])(nil)

// handleFunc is a function that handles the data of a topic
type HandleFunc[T input] func(context.Context, T) error

// handle is a struct that holds the name and the function of a handle
type handle[T input] struct {
	name        string
	handleFunc  HandleFunc[T]
	middlewares []middlewareFunc[T]
	metadata    metadata
}

func (h *handle[T]) Metadata(key string) (interface{}, bool) {
	return h.metadata.Get(key)
}

func (h *handle[T]) Next(ctx context.Context, data T) error {
	return h.handleFunc(ctx, data)
}

func (h *handle[T]) Use(m ...middlewareFunc[T]) {
	h.middlewares = append(h.middlewares, m...)
}

func (h *handle[T]) Name() string {
	return h.name
}

func (h *handle[T]) HandleFunc() interface{} {
	return h.handleFunc
}

func (h *handle[T]) run(ctx context.Context, data T, c *PubsubConfig) error {
	handleFunc := h.handleFunc

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handleFunc = h.middlewares[i](h, c)
	}

	return handleFunc(h.setupPublishContext(ctx), data)
}

func (h *handle[T]) setupPublishContext(c context.Context) context.Context {
	ctx := context.WithValue(c, "name", h.name)

	return ctx
}

// newHandler creates a new handle
func newHandler[T input](handleFunc HandleFunc[T], cfg *PubsubConfig) *handle[T] {
	return &handle[T]{
		name:        cfg.name,
		handleFunc:  handleFunc,
		middlewares: make([]middlewareFunc[T], 0),
		metadata:    cfg.metadata,
	}
}

// middlewareFunc is a function that wraps a handleFunc
type middlewareFunc[T input] func(Handler[T], *PubsubConfig) HandleFunc[T]

// type middlewareFunc[T input] func(*handle[T]) HandleFunc[T]

type metadata map[string]interface{}

func (m metadata) Get(key string) (interface{}, bool) {
	val, ok := m[key]

	return val, ok
}

func (m metadata) Set(key string, val interface{}) {
	m[key] = val
}
