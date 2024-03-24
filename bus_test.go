package sbus

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	defer cleanup()

	if defaultBus != nil {
		t.Error("expected default bus to be nil")
	}

	b := New()

	if b == nil {
		t.Error("expected bus to be created")
	}

	if defaultBus == nil {
		t.Error("expected default bus to be set")
	}
}

func TestSubscribe(t *testing.T) {
	defer cleanup()

	b := New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	})

	if h, ok := b.getHandle("greeting"); !ok {
		t.Error("expected topic to exist")
	} else {
		if len(h) != 1 {
			t.Error("expected handler to be added")
		}

		if h[0].Name() != "topic: greeting" {
			t.Error("expected handler name to be topic: greeting")
		}
	}
}

func TestSubscribe_WithtBus(t *testing.T) {
	defer cleanup()

	New() // create default bus
	b := New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	}, WithBus(b))

	if _, ok := b.getHandle("greeting"); !ok {
		t.Error("expected topic to exist")
	}

	if _, ok := defaultBus.getHandle("greeting"); ok {
		t.Error("expected topic to not exist")
	}
}

func TestSubscribe_WithtName(t *testing.T) {
	defer cleanup()

	New()

	Subscribe("topic:greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	}, WithName("greeting"))

	handlers, ok := defaultBus.getHandle("topic:greeting")

	if !ok {
		t.Error("expected topic to exist")
	}

	for _, hndl := range handlers {
		if hndl.Name() != "greeting" {
			t.Error("expected handler name to be greeting")
		}
	}
}

func TestSubscribe_WithMetadata(t *testing.T) {
	defer cleanup()

	New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	}, WithMetadata("key", "value"))

	handlers, ok := defaultBus.getHandle("greeting")

	if !ok {
		t.Error("expected topic to exist")
	}

	for _, hndl := range handlers {
		if v, ok := hndl.Metadata("key"); !ok || v != "value" {
			t.Error("expected metadata to be set")
		}
	}
}

func TestSubscribe_Multiple(t *testing.T) {
	defer cleanup()

	New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello1, " + name)

		return nil
	}, WithName("greeting:1"))

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello2, " + name)

		return nil
	}, WithName("greeting:2"))

	Publish("greeting", context.Background(), "Benbe")

	if _, ok := defaultBus.getHandle("greeting"); !ok {
		t.Error("expected topic to exist")
	}
}

func TestPublish(t *testing.T) {
	defer cleanup()

	New()

	val := ""
	Subscribe("greeting", func(ctx context.Context, name string) error {
		val = name

		return nil
	})

	err := Publish("greeting", context.Background(), "Benbe")
	if err != nil {
		t.Error(err)
	}

	if val != "Benbe" {
		t.Error("expected val to be Benbe")
	}
}

func TestPublish_WithoutTopic(t *testing.T) {
	defer cleanup()

	New()

	err := Publish("greeting", context.Background(), "Benbe")
	if err == nil {
		t.Error("expected error to be returned")
	}
}

func TestPublish_WithMiddleware(t *testing.T) {
	defer cleanup()

	New()

	val := ""
	Subscribe("greeting", func(ctx context.Context, name string) error {
		val = name

		return nil
	}).Use(lowerMiddleware) // LogMiddleware, OtelMiddleware

	Publish("greeting", context.Background(), "Benbe")

	if val != "benbe" {
		t.Error("expected val to be benbe")
	}
}

func cleanup() {
	defaultBus = nil
}

func lowerMiddleware[T string](h Handler, c PubsubConfig) HandleFunc[T] {
	return func(ctx context.Context, data T) error {
		str, ok := any(data).(string)

		if !ok {
			return fmt.Errorf("expected data to be of type string")
		}

		lower := strings.ToLower(str)

		return h.Handle(ctx, T(lower))

		// return next(ctx, T(lower))
	}
}

func LogMiddleware[T any](next HandleFunc[T]) HandleFunc[T] {
	return func(ctx context.Context, data T) error {
		fmt.Printf("data: %v\n", data)

		return next(ctx, data)
	}
}
