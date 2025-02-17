package main

import (
	"fmt"
	"os"

	"github.com/DobryySoul/Calc-service/internal/agent/config"
	"github.com/DobryySoul/Calc-service/internal/orchestrator/application"
)

func main() {
	cfg, err := config.NewConfigForAgent()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app := application.NewApplication(cfg)
	app.Run()
}
