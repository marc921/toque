package main

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Env string `env:"ENV"`
	// RabbitMQ
	RabbitMQ struct {
		URL string `env:"RABBITMQ_URL"`
	}
}

func ParseConfig(ctx context.Context) (*Config, error) {
	cfg := new(Config)
	err := envconfig.Process(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}
