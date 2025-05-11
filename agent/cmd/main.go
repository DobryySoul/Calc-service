package main

import (
	"context"
	"fmt"
	"os"

	"agent/internal/application"
	"agent/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nCONFIG: %+v\n", cfg)

	app, err := application.NewApplicationAgent(cfg)
	if err != nil {
		panic(err)
	}

	exitCode := app.Run(context.Background())
	os.Exit(exitCode)
}
