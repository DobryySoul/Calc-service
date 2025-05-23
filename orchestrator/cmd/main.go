package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DobryySoul/orchestrator/internal/application"
	"github.com/DobryySoul/orchestrator/internal/config"
)

func main() {
	cfg, err := config.LoadConfigEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	app := application.NewApplicationOrchestrator(cfg)
	app.Run(ctx)
}
