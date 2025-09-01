package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ose-micro/core/domain"
	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
)

type Config struct {
	Url          string        `mapstructure:"url"`
	Name         string        `mapstructure:"name"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	Timeout      time.Duration `mapstructure:"timeout"`
	MaxReconnect int           `mapstructure:"max_reconnect"`
}

type natsBus struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	log    logger.Logger
	tracer tracing.Tracer
}

func New(conf Config, log logger.Logger, tracer tracing.Tracer) (domain.Bus, error) {

	if conf.Name == "" {
		return nil, fmt.Errorf("nats bus name is required")
	}

	if conf.User == "" {
		return nil, fmt.Errorf("nats bus user is required")
	}

	if conf.Password == "" {
		return nil, fmt.Errorf("nats bus password is required")
	}

	if conf.Timeout == 0 {
		conf.Timeout = 30 * time.Second
	}

	if conf.MaxReconnect == 0 {
		conf.MaxReconnect = 5
	}

	opts := []nats.Option{
		nats.Name(conf.Name),
		nats.UserInfo(conf.User, conf.Password),
		nats.ReconnectWait(conf.Timeout),
		nats.MaxReconnects(conf.MaxReconnect),
	}

	nc, err := nats.Connect(conf.Url, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	return &natsBus{
		nc:     nc,
		js:     js,
		log:    log,
		tracer: tracer,
	}, nil
}
