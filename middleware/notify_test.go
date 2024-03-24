package middleware

import (
	"context"
	"testing"

	"github.com/loopilot/sbus"
)

func TestNotify(t *testing.T) {
	sbus.New()

	val := ""
	sbus.Subscribe("greeting", func(ctx context.Context, name string) error {
		val = name

		return nil
	}).Use(MiddlewareNotify)

	notify := &Notify{}

	err := sbus.Publish("greeting", context.Background(), "Benbe", WithNotify(notify))
	if err != nil {
		t.Error(err)
	}

	if val != "Benbe" {
		t.Error("expected val to be Benbe")
	}

	// t.Fail()
}
