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

	app := application.NewApplicationAgent(cfg)
	exitCode := app.Run(context.Background())
	os.Exit(exitCode)
}
