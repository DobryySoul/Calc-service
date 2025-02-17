package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DobryySoul/Calc-service/internal/agent/application"
	"github.com/DobryySoul/Calc-service/internal/agent/config"
)

func main() {
	cfg, err := config.NewConfigForAgent()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	app := application.NewApplicationAgent(cfg)
	app.Run(ctx)
}
