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

func (a *Application) Run(ctx context.Context) {
	defer close(a.tasks)
	defer close(a.results)

	for i := 0; i < a.cfg.ComputingPOWER; i++ {
		go worker(a.tasks, a.results, a.ready)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.ready:
			go func() {
				task := a.client.GetTask()
				if task == nil {
					a.ready <- struct{}{}
				}
				a.tasks <- *task
			}()
		case result := <-a.results:
			a.client.SendResult(result)
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
