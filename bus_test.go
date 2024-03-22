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

	if _, ok := b.getTopic("greeting"); !ok {
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

	if _, ok := b.getTopic("greeting"); !ok {
		t.Error("expected topic to exist")
	}

	if _, ok := defaultBus.getTopic("greeting"); ok {
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

	handlers, ok := defaultBus.getTopic("topic:greeting")

	if !ok {
		t.Error("expected topic to exist")
	}

	for _, hndl := range handlers {
		if hndl.Name() != "greeting" {
			t.Error("expected handler name to be greeting")
		}
	}

	// else {
	// 	h, ok := handlers[0].(*handle[string])

	// 	if !ok {
	// 		t.Error("expected handler to be of type handle[string]")
	// 	}

	// 	if h.name != "greeting" {
	// 		t.Error("expected handler name to be greeting")
	// 	}
	// }
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

	if _, ok := defaultBus.getTopic("greeting"); !ok {
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
	}).Use(lowerMiddleware, LogMiddleware, OtelMiddleware)

	Publish("greeting", context.Background(), "Benbe")

	if val != "benbe" {
		t.Error("expected val to be benbe")
	}
}

func cleanup() {
	defaultBus = nil
}

func lowerMiddleware[T string](next HandleFunc[T]) HandleFunc[T] {
	return func(ctx context.Context, data T) error {
		str, ok := any(data).(string)

		if !ok {
			return fmt.Errorf("expected data to be of type string")
		}

		lower := strings.ToLower(str)

		return next(ctx, T(lower))
	}
}

func LogMiddleware[T any](next HandleFunc[T]) HandleFunc[T] {
	return func(ctx context.Context, data T) error {
		fmt.Printf("data: %v\n", data)

		return next(ctx, data)
	}
}
