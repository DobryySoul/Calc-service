package main

import (
	"context"
	"os"

	"agent/internal/application"
	"agent/internal/config"
)

func main() {
	cfg := config.NewConfig()

	app := application.NewApplicationAgent(cfg)
	exitCode := app.Run(context.Background())
	os.Exit(exitCode)
}
