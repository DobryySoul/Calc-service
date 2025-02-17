package application

import (
	"github.com/DobryySoul/Calc-service/internal/orchestrator/config"
)

type Application struct {
	cfg     *config.Config
	client  *client.Client
	tasks   chan task.Task
	results chan result.Result
	ready   chan struct{}
}

func NewApplication(cfg *config.Config) *Application {
	return &Application{
		cfg: cfg,
		
	}
}
