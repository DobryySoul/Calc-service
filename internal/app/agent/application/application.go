package application

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/DobryySoul/Calc-service/internal/app/agent/config"
	"github.com/DobryySoul/Calc-service/internal/http/client"
	"github.com/DobryySoul/Calc-service/internal/http/models"
)

type Application struct {
	cfg     config.Config
	client  *client.Client
	tasks   chan models.Task
	results chan models.Result
	ready   chan struct{}
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

func NewApplicationAgent(cfg *config.Config) *Application {
	return &Application{
		cfg:     *cfg,
		client:  &client.Client{Host: cfg.Host, Port: cfg.Port},
		tasks:   make(chan models.Task),
		results: make(chan models.Result),
		ready:   make(chan struct{}, cfg.ComputingPOWER),
	}
}

func (app *Application) Run(ctx context.Context) int {
	var wg sync.WaitGroup

	defer close(app.results)
	defer close(app.tasks)
	defer wg.Wait()

	for i := 0; i < app.cfg.ComputingPOWER; i++ {
		wg.Add(1)
		go runWorker(app.tasks, app.results, app.ready, &wg)
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

func runWorker(tasks <-chan models.Task, results chan<- models.Result, ready chan<- struct{}, wg *sync.WaitGroup) {
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
			results <- models.Result{
				ID:    task.ID,
				Value: "Некорректные аргументы",
			}
		} else {
			value := ops[task.Operation](arg1, arg2)
			results <- models.Result{
				ID:    task.ID,
				Value: fmt.Sprintf("%f", value),
			}
		}
	}
}
