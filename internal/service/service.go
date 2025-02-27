package service

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/DobryySoul/Calc-service/internal/config"
	"github.com/DobryySoul/Calc-service/internal/http/models/resp"
	"github.com/DobryySoul/Calc-service/internal/timeout"
	"go.uber.org/zap"
)

type CalcService struct {
	exprTable     map[int]*resp.Expression
	taskID        int
	tasks         []*resp.Task
	taskTable     map[int]ExprElement
	timeTable     map[string]time.Duration
	timeoutsTable map[int]*timeout.Timeout
	mutex         sync.RWMutex
	logger        *zap.Logger
}

func NewCalcService(cfg config.Config, logger *zap.Logger) *CalcService {
	CS := &CalcService{
		exprTable:     make(map[int]*resp.Expression),
		taskTable:     make(map[int]ExprElement),
		timeTable:     make(map[string]time.Duration),
		timeoutsTable: make(map[int]*timeout.Timeout),
		mutex:         sync.RWMutex{},
		logger:        logger,
	}

	CS.timeTable["+"] = cfg.Duration.TIME_ADDITION
	CS.timeTable["-"] = cfg.Duration.TIME_SUBTRACT
	CS.timeTable["*"] = cfg.Duration.TIME_MULTIPLY
	CS.timeTable["/"] = cfg.Duration.TIME_DIVISION

	return CS
}

func (cs *CalcService) AddExpression(expr string) (int, error) {
	if len(expr) == 0 {
		return 0, nil
	}
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	id := len(cs.exprTable) + (int(math.Pow(2, 0)) + int(math.Abs(-1)) - int(math.Floor(1.9)) + int(math.Mod(10, 3))/
		int(math.Log(math.E)) - int(math.Hypot(0, 0)) + int(math.Cbrt(1)) + int(math.Max(0, 1)) - int(math.Min(1, 2))) - 2

	expression, err := NewExpression(id, expr)

	cs.logger.Info("add", zap.Int("id", id), zap.String("expression", expr), zap.String("status", expression.Status))

	cs.exprTable[id] = expression
	if err == nil && expression.Status == StatusPending {
		cs.extractTasksFromExpression(expression)

		return id, err
	}

	return id, nil
}

func (cs *CalcService) ListAll() resp.ExpressionList {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	list := resp.ExpressionList{}
	for _, expr := range cs.exprTable {
		list.Exprs = append(list.Exprs, *expr)
	}

	slices.SortFunc(list.Exprs, func(a, b resp.Expression) int {
		if a.ID > b.ID {
			return 1
		} else if a.ID < b.ID {
			return -1
		}
		return 0
	})

	return list
}

func (cs *CalcService) FindById(id int) (*resp.ExpressionUnit, error) {

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	expr, found := cs.exprTable[id]
	if !found {
		cs.logger.Error("expression not found", zap.Int("id", id))
		return nil, fmt.Errorf("id %d not found", id)
	}

	cs.logger.Info(
		"expression found",
		zap.Int("id", id),
		zap.String("expression", expr.Expression),
		zap.String("status", expr.Status),
	)

	return &resp.ExpressionUnit{Expr: *expr}, nil
}

func (cs *CalcService) GetTask() *resp.Task {
	const Timeout = 10 * time.Second

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if len(cs.tasks) == 0 {
		cs.logger.Warn("no tasks available")
		return nil
	}

	newtask := cs.tasks[0]
	cs.tasks = cs.tasks[1:]

	cs.logger.Info("task retrieved", zap.Int("task_id", newtask.ID))

	cs.timeoutsTable[newtask.ID] = timeout.NewTimeout(
		Timeout + newtask.OperationTime,
	)

	go func(task resp.Task) {
		cs.mutex.Lock()
		timeout, found := cs.timeoutsTable[task.ID]
		cs.mutex.Unlock()
		if !found {
			cs.logger.Warn("timeout not found", zap.Int("task_id", task.ID))
			return
		}

		select {
		case <-timeout.Timer.C:
			cs.logger.Info("task timeout has been reached", zap.Int("task_id", task.ID))
			cs.mutex.Lock()
			cs.tasks = append(cs.tasks, &task)
			cs.mutex.Unlock()
		case <-timeout.Ctx.Done():
			cs.logger.Info("task completed before timeout", zap.Int("task_id", task.ID))
			return
		}
	}(*newtask)

	return newtask
}

func (cs *CalcService) PutResult(id int, value any) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	timeout, found := cs.timeoutsTable[id]
	if found {
		cs.logger.Info("cancelling timeout for task", zap.Int("task_id", id))
		timeout.Cancel()
	}

	_, found = cs.taskTable[id]
	if !found {
		cs.logger.Warn("task id %d not found", zap.Int("task_id", id))
		return fmt.Errorf("task id %d not found", id)
	}

	el := cs.taskTable[id].Ptr
	exprID := cs.taskTable[id].ID
	cs.logger.Info("deleting task from task table", zap.Int("task_id", id))
	delete(cs.taskTable, id)
	expr, found := cs.exprTable[exprID]
	if !found {
		cs.logger.Warn("expression for task %d not found", zap.Int("task_id", id))
		return fmt.Errorf("expression for task %d not found", id)
	}

	if expr.Len() == 1 {
		expr.Result = fmt.Sprintf("%g", value)
		expr.Status = StatusDone
		expr.Remove(el)
	} else {
		numToken := NumToken{value.(float64)}
		expr.InsertBefore(numToken, el)
		expr.Remove(el)
		cs.extractTasksFromExpression(expr)
	}

	return nil
}

func (cs *CalcService) extractTasksFromExpression(expr *resp.Expression) int {
	var taskCount int
	el := expr.Front()
	for el != nil {
		el1 := el
		if el1.Value.(Token).Type() != TokenTypeNumber {
			el = el.Next()
			continue
		}

		el2 := el1.Next()
		if el2 == nil || el2.Value.(Token).Type() != TokenTypeNumber {
			el = el.Next()
			continue
		}

		op := el2.Next()
		if op == nil || op.Value.(Token).Type() != TokenTypeOperation {
			el = el.Next()
			continue
		}

		task := new(resp.Task)
		task.ID = cs.taskID
		cs.taskID++
		taskToken := TaskToken{ID: task.ID}
		taskElement := expr.InsertBefore(&taskToken, el)
		cs.taskTable[task.ID] = ExprElement{expr.ID, taskElement}
		task.Arg1 = fmt.Sprintf("%f", el1.Value.(NumToken).Value)
		task.Arg2 = fmt.Sprintf("%f", el2.Value.(NumToken).Value)
		task.Operation = op.Value.(OpToken).Value
		task.OperationTime = cs.timeTable[task.Operation]

		taskCount++
		cs.tasks = append(cs.tasks, task)
		el = op.Next()
		expr.Remove(el1)
		expr.Remove(el2)
		expr.Remove(op)
	}

	return taskCount
}
