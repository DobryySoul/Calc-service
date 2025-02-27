package application

import (
	"context"
	"os"
	"os/signal"

	"github.com/DobryySoul/Calc-service/internal/config"
	"github.com/DobryySoul/Calc-service/internal/http/server"
	"github.com/DobryySoul/Calc-service/pkg/logger"
	"go.uber.org/zap"
)

type Application struct {
	cfg    config.Config
	logger *zap.Logger
}

func NewApplicationOrchestrator(cfg *config.Config) *Application {
	logger := logger.SetupLogger()
	return &Application{cfg: *cfg, logger: logger}
}

func (a *Application) Run(ctx context.Context) int {
	defer a.logger.Sync()

	shutDownFunc, err := server.Run(ctx, a.logger, a.cfg)
	if err != nil {
		a.logger.Error("Run server error", zap.String("error", err.Error()))
		return 1
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-c
	shutDownFunc(ctx)

	a.logger.Info("Server has been shut down")

	return 0
}
