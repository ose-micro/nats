# NATS Bus

> A wrapper around NATS JetStream
providing simple helpers for publishing, subscribing, and stream management.

## Features

 - Publish data as []byte, string, or any Go struct (auto JSON marshal). 
 - EnsureStream to create streams on the fly if they don’t exist. 
 - Subscribe with durable consumers, manual ack, and error handling. 
 - Graceful shutdown with connection draining and close.

## Installation
```go
go get github.com/your-org/your-nats-bus
```
## Quick Start
### Setup
```go
package main

import (
	"context"
	"time"
	"log"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
)

func main() {
	logs, err := logger.NewZap(logger.Config{})
	if err != nil {
		log.Fatal(err)
	}

	tracer, err := tracing.NewOtel(tracing.Config{
		Endpoint:    "nats://localhost:4222",
		ServiceName: "Nats",
		SampleRatio: 1.0,
	}, logs)
	if err != nil {
		log.Fatal(err)
	}

	bus, err := New(Config{
		Url:          "nats://turntable.proxy.rlwy.net:58598",
		Name:         "Ose Nats",
		User:         "",
		Password:     "",
		Timeout:      2 * time.Second,
		MaxReconnect: 5,
	}, log, tracer)
	if err != nil {
		log.Fatal(err)
	}

	events := []string{"events.*"}

	if err = bus.EnsureStream("EVENTS", events...); err != nil {
		log.Fatal(err)
	}

	if err := bus.Publish("events.created", map[string]any{
		"id":   "123",
		"name": "New Event",
	}); err != nil {
		log.Fatal(err)
	}

	if err := bus.Subscribe("events.created", "ose", func(ctx context.Context, data any) error {
		log.Printf("📩 received event: %#v", data)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err = bus.Close(); err != nil {
		log.Fatal(err)
	}
}

```
## API
```go
Publish(subject string, data any) error
```

Publish a message to a subject. <br />
Supports `[]byte`, `string`, or any struct (auto-marshaled to JSON).

---

```go
EnsureStream(name string, subjects ...string) error
```
EnsureStream(name string, subjects ...string) error <br/>
Ensure a JetStream stream exists with given name and subjects.
If the stream already exists, it does nothing.
---
`Subscribe(subject string, durable string, handler func(ctx context.Context, data any) error) error`

Subscribe to a subject with a durable consumer.
- Messages are JSON-unmarshaled if possible, otherwise passed as string.
- Successful handler → message is Ack()ed.
- Handler error → message is Nak()ed for retry.
---

`Close() error`

Gracefully drain and close the NATS connection.

---

## Example Output
```bash
✅ Created stream EVENTS with subjects [events.*]
INFO  Subscribed to subject subject=events.created durable=worker-1
📩 received event: map[string]interface {}{"id":"123", "name":"New Event"}
```