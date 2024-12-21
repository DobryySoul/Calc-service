package main

import (
	"log/slog"

	"github.com/DobryySoul/Calc-service/internal/application"
)

func main() {
	logger := slog.Default()

	logger.Info("starting server on", slog.String("PORT", application.ConfigFromEnv().Addr))

	app := application.New()

	logger.Info("server started", slog.String("addr", "localhost:"+application.ConfigFromEnv().Addr))
	// app.Run()

	app.RunServer()
}
