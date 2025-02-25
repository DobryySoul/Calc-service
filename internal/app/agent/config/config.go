package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ComputingPOWER int
	Host           string
	Port           int
}

func ParseFlags() (*string, *int) {
	
	var (
		host = flag.String("h", "localhost", "The host name of the orchestrator")
		port = flag.Int("p", 8081, "Port of the orchestrator")
	)
	
	flag.Parse()

	return host, port
}

func NewConfigForAgent() (*Config, error) {
	power, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || power < 0 {
		return nil, fmt.Errorf("failed to parse COMPUTING_POWER: %w", err)
	}

	host, port := ParseFlags()

	cfg := &Config{
		ComputingPOWER: power,
		Host:           *host,
		Port:           *port,
	}

	return cfg, nil
}
