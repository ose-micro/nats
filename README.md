# NATS + JetStream

This is a minimal setup for running NATS with JetStream enabled on Railway using Docker Compose.

## Features

- JetStream enabled with data persistence
- Monitoring endpoint on port 8222
- Debug logging
- Health check

## Deployment

1. Push this repo to GitHub.
2. Create a new project on [Railway](https://railway.app).
3. Link the GitHub repo.
4. Railway will detect the `docker-compose.yml` and deploy NATS.

## Ports

- `4222`: NATS Client
- `8222`: HTTP Monitoring (`http://<railway-url>:8222`)

## Connecting from Go (example)

```go
import (
	"github.com/nats-io/nats.go"
	"log"
)

func main() {
	nc, err := nats.Connect("nats://<your-app>.up.railway.app:4222")
	if err != nil {
		log.Fatal(err)
	}
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	js.Publish("test.subject", []byte("Hello, JetStream!"))
}
