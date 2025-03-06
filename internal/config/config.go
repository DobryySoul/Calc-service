package config

import (
	"fmt"
	"time"

	"github.com/goloop/env"
)

type Config struct {
	Host           string `env:"HOST" default:"localhost"`
	Port           string `env:"PORT" default:"9090"`
	ComputingPOWER int    `env:"COMPUTING_POWER" default:"2"`
	TIME_ADDITION  time.Duration
	TIME_SUBTRACT  time.Duration
	TIME_MULTIPLY  time.Duration
	TIME_DIVISION  time.Duration
}

type Time struct {
	TIME_ADDITION string `env:"TIME_ADDITION_MS" default:"2000"`
	TIME_SUBTRACT string `env:"TIME_SUBTRACTION_MS" default:"2000"`
	TIME_MULTIPLY string `env:"TIME_MULTIPLICATIONS_MS" default:"4000"`
	TIME_DIVISION string `env:"TIME_DIVISIONS_MS" default:"4000"`
}

func LoadConfigEnv() (*Config, error) {
	const filePath = ".env"

	if err := env.Load(filePath); err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	var cfg Config

	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}
	var Time Time
	if err := env.Unmarshal("", &Time); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}

	cfg.TIME_ADDITION, _ = time.ParseDuration(Time.TIME_ADDITION + "ms")
	cfg.TIME_SUBTRACT, _ = time.ParseDuration(Time.TIME_SUBTRACT + "ms")
	cfg.TIME_MULTIPLY, _ = time.ParseDuration(Time.TIME_MULTIPLY + "ms")
	cfg.TIME_DIVISION, _ = time.ParseDuration(Time.TIME_DIVISION + "ms")

	return &cfg, nil
}
