package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func (n *natsBus) Publish(subject string, data any) error {
	var payload []byte
	var err error

	switch v := data.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		payload, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("publish marshal: %w", err)
		}
	}

	_, err = n.js.Publish(subject, payload)
	return err
}

func (n *natsBus) EnsureStream(name string, subjects ...string) error {
	if name == "" {
		return fmt.Errorf("stream name required")
	}
	if len(subjects) == 0 {
		return fmt.Errorf("at least one subject required")
	}

	// Check if stream exists
	_, err := n.js.StreamInfo(name)
	if err == nil {
		return nil
	}

	// Create stream
	_, err = n.js.AddStream(&nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   nats.FileStorage,
		Replicas:  1,
		Retention: nats.LimitsPolicy,
		MaxMsgs:   -1,
		MaxBytes:  -1,
	})
	if err != nil {
		return fmt.Errorf("add stream: %w", err)
	}

	log.Printf("✅ Created stream %s with subjects %v", name, subjects)
	return nil
}

func (n *natsBus) Subscribe(subject string, durable, queue string, handler func(ctx context.Context, data any) error) error {
	var err error
	var sub *nats.Subscription

	if queue != "" {
		sub, err = n.js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			ctx := context.Background()

			var data any
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				data = string(msg.Data)
			}

			if herr := handler(ctx, data); herr == nil {
				_ = msg.Ack()
			} else {
				_ = msg.Nak()
			}
		},
			nats.Durable(durable),
			nats.ManualAck(),
			nats.AckWait(30*time.Second),
			nats.MaxAckPending(1000),
		)
	} else {
		sub, err = n.js.Subscribe(subject, func(msg *nats.Msg) {
			ctx := context.Background()

			var data any
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				data = string(msg.Data)
			}

			if herr := handler(ctx, data); herr == nil {
				_ = msg.Ack()
			} else {
				_ = msg.Nak()
			}
		},
			nats.Durable(durable),
			nats.ManualAck(),
			nats.AckWait(30*time.Second),
			nats.MaxAckPending(1000),
		)
	}
	if err != nil {
		n.log.Error("Fail to subscribe", "error", err)
		return err
	}

	// Flush subscription
	if err := n.nc.Flush(); err != nil {
		n.log.Error("NATS flush failed", "error", err)
		return err
	}
	if lastErr := n.nc.LastError(); lastErr != nil {
		n.log.Error("NATS connection error after subscribe", "error", lastErr)
		return lastErr
	}

	n.log.Info("Subscribed to subject", "subject", subject, "durable", durable, "queue", queue)
	_ = sub

	return nil
}

func (n *natsBus) Close() error {
	if n.nc != nil && !n.nc.IsClosed() {
		err := n.nc.Drain()
		if err != nil {
			return err
		}

		n.nc.Close()
	}

	return nil
}
