package application

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/DobryySoul/Calc-service/internal/configs"
	"github.com/DobryySoul/Calc-service/internal/http/client"
	"github.com/DobryySoul/Calc-service/internal/http/models/req"
	"github.com/DobryySoul/Calc-service/internal/http/models/resp"
	"github.com/DobryySoul/Calc-service/pkg/logger"
	"go.uber.org/zap"
)

type Application struct {
	cfg     configs.Config
	client  *client.Client
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

func NewApplicationAgent(cfg *configs.Config) *Application {
	logger := logger.SetupLogger()

	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		panic(err)
	}
	return &Application{
		cfg:     *cfg,
		client:  &client.Client{Host: cfg.Host, Port: port},
		tasks:   make(chan resp.Task),
		results: make(chan req.Result),
		ready:   make(chan struct{}, cfg.ComputingPOWER),
		wg:      sync.WaitGroup{},
		logger:  logger,
	}
}

func (app *Application) Run(ctx context.Context) int {
	defer close(app.results)
	defer close(app.tasks)
	defer app.wg.Wait()

	for i := 1; i <= app.cfg.ComputingPOWER; i++ {
		app.wg.Add(1)
		app.logger.Info("Worker has been started", zap.Int("worker", i))
		go runWorker(app.tasks, app.results, app.ready, &app.wg)
	}

	for {
		select {
		case <-ctx.Done():
			return 0
		case res := <-app.results:
			app.client.SendResult(res)
		case <-app.ready:
			task := app.client.GetTask()
			if task == nil {
				app.ready <- struct{}{}
			} else {
				app.tasks <- *task
			}
		}
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
				ID:    task.ID,
				Value: "Некорректные аргументы",
			}
		} else {
			value := ops[task.Operation](arg1, arg2)
			results <- req.Result{
				ID:    task.ID,
				Value: fmt.Sprintf("%f", value),
			}
		}
	}
}
