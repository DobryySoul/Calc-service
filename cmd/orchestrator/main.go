package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DobryySoul/Calc-service/internal/orchestrator/application"
	"github.com/DobryySoul/Calc-service/internal/orchestrator/config"
)

func main() {
	cfg, err := config.NewConfigForOrchestrator()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	app := application.NewApplicationOrchestrator(cfg)
	app.Run(ctx)
}
