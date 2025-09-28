package nats_test

import (
	"context"
	"testing"
	"time"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"github.com/ose-micro/nats"
)

func TestPublish(t *testing.T) {
	log, err := logger.NewZap(logger.Config{})
	if err != nil {
		t.Error(err)
	}

	tracer, err := tracing.NewOtel(tracing.Config{
		Endpoint:    "nats://localhost:4222",
		ServiceName: "Nats",
		SampleRatio: 1.0,
	}, log)
	if err != nil {
		t.Error(err)
	}

	bus, err := nats.New(nats.Config{
		Url:          "nats://localhost:4222",
		Name:         "ose",
		User:         "moriba",
		Password:     "yT7f9hK3vLqP8dX2bHcM",
		Timeout:      2 * time.Second,
		MaxReconnect: 5,
	}, log, tracer)
	if err != nil {
		t.Error(err)
	}

	events := []string{"events.*"}

	err = bus.EnsureStream("TEST", events...)
	if err != nil {
		t.Error(err)
	}

	err = bus.Publish("events.created", events)
	if err != nil {
		t.Error(err)
	}

	err = bus.Subscribe("events.created", "ose-queue", "ose", func(ctx context.Context, data any) error {
		t.Log(data)

		return nil
	})
	if err != nil {
		t.Error(err)
	}

	err = bus.Close()
	if err != nil {
		t.Error(err)
	}
}
