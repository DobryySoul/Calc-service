package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DobryySoul/Calc-service/internal/app/orchestrator/application"
	"github.com/DobryySoul/Calc-service/internal/app/orchestrator/config"
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
