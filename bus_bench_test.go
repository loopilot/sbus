package sbus

import (
	"context"
	"testing"
)

func init() {

}

func BenchmarkPublish(b *testing.B) {
	New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		return nil
	})

	b.StopTimer()
	b.ResetTimer()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := Publish("greeting", context.Background(), "John")

		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkPublish_WithBus(b *testing.B) {
	bus := New()

	Subscribe("greeting", func(ctx context.Context, name string) error {
		return nil
	}, WithBus(bus))

	b.StopTimer()
	b.ResetTimer()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Publish("greeting", context.Background(), "John", WithBus(bus))
	}
}
