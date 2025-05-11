package config

import (
	"fmt"
	"time"

	"github.com/goloop/env"
)

type Config struct {
	Host           string `env:"HOST" default:"orchestrator"`
	Port           string `env:"PORT" default:"9090"`
	GRPCPort       string `env:"GRPC_PORT" default:"50051"`
	ComputingPOWER int    `env:"COMPUTING_POWER" default:"3"`
	PostgresConfig PostgresConfig
	JWTConfig      JWTConfig
	TIME_ADDITION  time.Duration
	TIME_SUBTRACT  time.Duration
	TIME_MULTIPLY  time.Duration
	TIME_DIVISION  time.Duration
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" default:"postgres"`
	Port     string `env:"POSTGRES_PORT" default:"5432"`
	Username string `env:"POSTGRES_USERNAME" default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" default:"05042007PULlup!"`
	Database string `env:"POSTGRES_DATABASE" default:"postgres"`
	MaxConns int    `env:"POSTGRES_MAX_CONN" default:"15"`
	MinConns int    `env:"POSTGRES_MIN_CONN" default:"0"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" default:"secret"`
	TTL    string `env:"JWT_TTL" default:"2h"`
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

	var PostgresConfig PostgresConfig
	if err := env.Unmarshal("", &PostgresConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}

	var JWTConfig JWTConfig
	if err := env.Unmarshal("", &JWTConfig); err != nil {
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

	cfg.PostgresConfig = PostgresConfig
	cfg.JWTConfig = JWTConfig

	return &cfg, nil
}
