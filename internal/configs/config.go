package configs

import (
	"fmt"
	"time"

	"github.com/goloop/env"
)

type Config struct {
	Host           string `env:"HOST" default:"localhost"`
	Port           string `env:"PORT" default:"8080"`
	ComputingPOWER int    `env:"COMPUTING_POWER" default:"2"`
	Duration       Duration
}

type Duration struct {
	TIME_ADDITION time.Duration `env:"TIME_ADDITION_MS" default:"2000"`
	TIME_SUBTRACT time.Duration `env:"TIME_SUBTRACTION_MS" default:"2000"`
	TIME_MULTIPLY time.Duration `env:"TIME_MULTIPLICATIONS_MS" default:"2000"`
	TIME_DIVISION time.Duration `env:"TIME_DIVISIONS_MS" default:"2000"`
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

	return &cfg, nil
}
