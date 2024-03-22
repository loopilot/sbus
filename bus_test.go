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

	if _, ok := b.GetTopic("greeting"); !ok {
		t.Error("expected topic to exist")
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

	if _, ok := b.GetTopic("greeting"); !ok {
		t.Error("expected topic to exist")
	}

	if _, ok := defaultBus.GetTopic("greeting"); ok {
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

	if handlers, ok := defaultBus.GetTopic("topic:greeting"); !ok {
		t.Error("expected topic to exist")
	} else {
		h, ok := handlers[0].(*handler[string])

		if !ok {
			t.Error("expected handler to be of type handler[string]")
		}

		if h.name != "greeting" {
			t.Error("expected handler name to be greeting")
		}
	}
}

func TestSubscribe_Multiple(t *testing.T) {
	defer cleanup()

	New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	}, WithName("greeting:1"))

	Subscribe("greeting", func(ctx context.Context, name string) error {
		fmt.Println("Hello, " + name)

		return nil
	}, WithName("greeting:2"))

	if handlers, ok := defaultBus.GetTopic("greeting"); !ok {
		t.Error("expected topic to exist")
	} else {
		for i, hndl := range handlers {
			h, ok := hndl.(*handler[string])

			if !ok {
				t.Error("expected handler to be of type handler[string]")
			}

			if h.name != fmt.Sprintf("greeting:%d", i+1) {
				t.Error("expected handler name to be greeting")
			}
		}
	}
}

func TestPublish(t *testing.T) {
	defer cleanup()

	New()

	val := "Benbe"
	Subscribe("greeting", func(ctx context.Context, name string) error {
		val = name

		return nil
	})

	Publish("greeting", context.Background(), "Benbe")

	if val != "Benbe" {
		t.Error("expected val to be Benbe")
	}
}

func TestPublish_WithMiddleware(t *testing.T) {
	defer cleanup()

	New()

	val := ""
	Subscribe("greeting", func(ctx context.Context, name string) error {
		val = name

		return nil
	}).Use(lowerMiddleware)

	Publish("greeting", context.Background(), "Benbe")

	if val != "benbe" {
		t.Error("expected val to be benbe")
	}
}

func cleanup() {
	defaultBus = nil
}

func lowerMiddleware[T string](next handlerFunc[T]) handlerFunc[T] {
	return func(ctx context.Context, data T) error {
		str, ok := any(data).(string)

		if !ok {
			return fmt.Errorf("expected data to be of type string")
		}

		lower := strings.ToLower(str)

		return next(ctx, T(lower))
	}
}
