package application

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/DobryySoul/Calc-service/internal/agent/config"
	"github.com/DobryySoul/Calc-service/internal/http/client"
	"github.com/DobryySoul/Calc-service/internal/result"
	"github.com/DobryySoul/Calc-service/internal/task"
)

var ops map[string]func(float64, float64) float64

func init() {
	ops = map[string]func(float64, float64) float64{
		"+": func(a, b float64) float64 { return a + b },
		"-": func(a, b float64) float64 { return a - b },
		"*": func(a, b float64) float64 { return a * b },
		"/": func(a, b float64) float64 { return a / b },
	}
}

type Application struct {
	cfg     *config.Config
	client  *client.Client
	tasks   chan task.Task
	results chan result.Result
	ready   chan struct{}
}

func NewApplicationAgent(cfg *config.Config) *Application {

	return &Application{
		cfg:     cfg,
		client:  &client.Client{Host: cfg.Host, Port: cfg.Port},
		tasks:   make(chan task.Task),
		results: make(chan result.Result),
		ready:   make(chan struct{}, cfg.ComputingPOWER),
	}
}

func (app *Application) Run(ctx context.Context) int {
	defer close(app.tasks)
	defer close(app.results)

	for i := 0; i < app.cfg.ComputingPOWER; i++ {
		go worker(app.tasks, app.results, app.ready)
	}

	for {
		select {
		case <-ctx.Done():
			return 0
		case <-app.ready:
			go func() {
				task := app.client.GetTask()
				if task == nil {
					app.ready <- struct{}{}
				} else {
					app.tasks <- *task
				}
			}()
		case result := <-app.results:
			app.client.SendResult(result)
		}
	}
}

func worker(tasks <-chan task.Task, results chan<- result.Result, ready chan<- struct{}) {
	for {
		ready <- struct{}{}
		task, ok := <-tasks
		if !ok {
			break
		}

		time.Sleep(task.OperationTime)

		arg1, err1 := strconv.ParseFloat(task.Arg1, 64)
		arg2, err2 := strconv.ParseFloat(task.Arg2, 64)
		if err1 != nil || err2 != nil {
			results <- result.Result{ID: task.ID, Value: fmt.Sprintf("%f", math.NaN)}
		} else {
			value := ops[task.Operation](arg1, arg2)
			results <- result.Result{ID: task.ID, Value: fmt.Sprintf("%f", value)}
		}
	}
}
