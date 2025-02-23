package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	TIME_ADDITION time.Duration
	TIME_SUBTRACT time.Duration
	TIME_MULTIPLY time.Duration
	TIME_DIVISION time.Duration
}

func NewConfigForOrchestrator() (*Config, error) {
	addition, err := time.ParseDuration(os.Getenv("TIME_ADDITION_MS") + "ms")
	if err != nil || addition < 0 {
		return nil, fmt.Errorf("failed to parse TIME_ADDITION_MS: %w", err)
	}
	subtract, err := time.ParseDuration(os.Getenv("TIME_SUBTRACTION_MS") + "ms")
	if err != nil || subtract < 0 {
		return nil, fmt.Errorf("failed to parse TIME_SUBTRACTION_MS: %w", err)
	}
	multiply, err := time.ParseDuration(os.Getenv("TIME_MULTIPLICATIONS_MS") + "ms")
	if err != nil || multiply < 0 {
		return nil, fmt.Errorf("failed to parse TIME_MULTIPLICATIONS_MS: %w", err)
	}
	division, err := time.ParseDuration(os.Getenv("TIME_DIVISIONS_MS") + "ms")
	if err != nil || division < 0 {
		return nil, fmt.Errorf("failed to parse TIME_DIVISIONS_MS: %w", err)
	}

	cfg := &Config{
		TIME_ADDITION: addition,
		TIME_SUBTRACT: subtract,
		TIME_MULTIPLY: multiply,
		TIME_DIVISION: division,
	}

	return cfg, nil
}
