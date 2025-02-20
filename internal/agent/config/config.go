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
	flag.Parse()

	var (
		host = flag.String("h", "localhost", "The host name of the orchestrator")
		port = flag.Int("p", 8080, "Port of the orchestrator")
	)

	if *host == "" {
		*host = "localhost"
	}
	if *port == 0 {
		*port = 8080
		fmt.Printf("Incorrect port %d, using default value 8080", *port)
	}

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
