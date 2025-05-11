package application

import (
	"agent/internal/config"
	"agent/internal/grpc/client"
	"agent/internal/models/req"
	"agent/internal/models/resp"
	"agent/pkg/logger"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Application struct {
	cfg     *config.Config
	client  *client.GRPCClient
	tasks   chan resp.Task
	results chan req.Result
	ready   chan struct{}
	wg      sync.WaitGroup
	logger  *zap.Logger
}

var ops map[string]func(float64, float64) float64

func init() {
	ops = make(map[string]func(float64, float64) float64)
	ops["+"] = addition
	ops["-"] = subtraction
	ops["*"] = multiplication
	ops["/"] = division
}

func addition(a, b float64) float64       { return a + b }
func subtraction(a, b float64) float64    { return a - b }
func multiplication(a, b float64) float64 { return a * b }
func division(a, b float64) float64       { return a / b }

func NewApplicationAgent(cfg *config.Config) (*Application, error) {
	logger := logger.SetupLogger()

	logger.Error("Starting agent with config:", zap.Any("config", cfg))

	grpcClient, err := client.NewGRPCClient(cfg.Host, cfg.Port, logger)
	if err != nil {
		logger.Error("failed to create gRPC client", zap.Error(err))
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &Application{
		cfg:     cfg,
		client:  grpcClient,
		tasks:   make(chan resp.Task),
		results: make(chan req.Result),
		ready:   make(chan struct{}, cfg.ComputingPOWER),
		wg:      sync.WaitGroup{},
		logger:  logger,
	}, nil
}

func (app *Application) Run(ctx context.Context) int {
	defer app.cleanup()
	defer app.wg.Wait()

	for i := 1; i <= app.cfg.ComputingPOWER; i++ {
		app.wg.Add(1)
		app.logger.Info("Worker has been started", zap.Int("worker", i))
		go runWorker(app.tasks, app.results, app.ready, &app.wg)
	}

	for {
		select {
		case <-ctx.Done():
			app.logger.Info("Application stopped by context")
			return 0
		case res := <-app.results:
			app.client.SendResult(res, res.UserID)
		case <-app.ready:
			task := app.client.GetTask()
			if task == nil {
				app.ready <- struct{}{}
				time.Sleep(1 * time.Second)
			} else {
				app.tasks <- *task
			}
		}
	}
}

func (app *Application) cleanup() {
	app.logger.Info("Cleaning up resources...")
	close(app.results)
	close(app.tasks)

	if err := app.client.Close(); err != nil {
		app.logger.Error("failed to close gRPC client", zap.Error(err))
	}
}

func runWorker(tasks <-chan resp.Task, results chan<- req.Result, ready chan<- struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		ready <- struct{}{}
		task, ok := <-tasks
		if !ok {
			return
		}

		time.Sleep(task.OperationTime)

		arg1, err1 := strconv.ParseFloat(task.Arg1, 64)
		arg2, err2 := strconv.ParseFloat(task.Arg2, 64)

		if err1 != nil || err2 != nil {
			results <- req.Result{
				ID:     task.ID,
				Value:  0.0,
				UserID: task.UserID,
			}
		} else {
			value := ops[task.Operation](arg1, arg2)
			results <- req.Result{
				ID:     task.ID,
				Value:  value,
				UserID: task.UserID,
			}
		}
	}
}
