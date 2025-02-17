package application

import (
	"context"

	"github.com/DobryySoul/Calc-service/internal/orchestrator/config"
)

type Application struct {
	cfg config.Config
}

func NewApplicationOrchestrator(cfg *config.Config) *Application {
	return &Application{cfg: *cfg}
}

func (a *Application) Run(ctx context.Context) {
	
}
