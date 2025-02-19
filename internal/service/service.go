package service

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/DobryySoul/Calc-service/internal/orchestrator/config"
	"github.com/DobryySoul/Calc-service/internal/task"
	"github.com/DobryySoul/Calc-service/internal/timeout"
)

type CalcService struct {
	taskID        int
	tasks         []*task.Task
	exprTable     map[string]*Expression
	taskTable     map[int]ExprElement
	timeTable     map[string]time.Duration
	timeoutsTable map[int]*timeout.Timeout
	locker        sync.RWMutex
}

func NewCalcService(cfg config.Config) *CalcService {
	CS := &CalcService{
		exprTable:     make(map[string]*Expression),
		taskTable:     make(map[int]ExprElement),
		timeTable:     make(map[string]time.Duration),
		timeoutsTable: make(map[int]*timeout.Timeout),
	}

	CS.timeTable["+"] = cfg.TIME_ADDITION
	CS.timeTable["-"] = cfg.TIME_SUBTRACT
	CS.timeTable["*"] = cfg.TIME_MULTIPLY
	CS.timeTable["/"] = cfg.TIME_SUBTRACT

	return CS
}

func (cs *CalcService) AddExpression(id int, expr string) error {
	if len(expr) == 0 {
		return fmt.Errorf("empty expression")
	}

	cs.locker.Lock()
	defer cs.locker.Unlock()
	ID := strconv.Itoa(id)
	if _, found := cs.exprTable[ID]; found {
		return fmt.Errorf("not a unique ID: %q", id)
	}

	expression, err := NewExpression(id, expr)
	cs.exprTable[ID] = expression
	if err == nil && expression.Status == StatusPending {
		// cs.extractTasksFromExpression(expression)
	}
	return nil
}

func (cs *CalcService) ListAll() ExpressionList {
	cs.locker.RLock()
	defer cs.locker.RUnlock()

	list := ExpressionList{}
	for _, expr := range cs.exprTable {
		list.Exprs = append(list.Exprs, *expr)
	}

	slices.SortFunc(list.Exprs, func(a, b Expression) int {
		if a.ID > b.ID {
			return 1
		} else if a.ID < b.ID {
			return -1
		}
		return 0
	})

	return list
}

func (cs *CalcService) FindById(id string) (*ExpressionUnit, error) {
	cs.locker.RLock()
	defer cs.locker.RUnlock()

	expr, found := cs.exprTable[id]
	if !found {
		return nil, fmt.Errorf("id %q not found", id)
	}
	return &ExpressionUnit{Expr: *expr}, nil
}

func (cs *CalcService) GetTask() *task.Task {
	cs.locker.Lock()
	defer cs.locker.Unlock()
	if len(cs.tasks) == 0 {
		return nil
	}

	newtask := cs.tasks[0]
	cs.tasks = cs.tasks[1:]

	cs.timeoutsTable[newtask.ID] = timeout.NewTimeout(
		5*time.Second + newtask.OperationTime,
	)

	go func(task task.Task) {
		cs.locker.Lock()
		timeout, found := cs.timeoutsTable[task.ID]
		cs.locker.Unlock()
		if !found {
			return
		}

		select {
		case <-timeout.Timer.C:
			cs.locker.Lock()
			cs.tasks = append(cs.tasks, &task)
			cs.locker.Unlock()
		case <-timeout.Ctx.Done():
			return
		}

	}(*newtask)

	return newtask
}

func (cs *CalcService) PutResult(id int, value float64) error {
	cs.locker.Lock()
	defer cs.locker.Unlock()

	timeout, found := cs.timeoutsTable[id]
	if found {
		timeout.Cancel()
	}

	_, found = cs.taskTable[id]
	if !found {
		return fmt.Errorf("task id %d not found", id)
	}

	el := cs.taskTable[id].Ptr
	exprID := cs.taskTable[id].ID
	delete(cs.taskTable, id)
	expr, found := cs.exprTable[strconv.Itoa(exprID)] // cs.exprTable[exprID]
	if !found {
		return fmt.Errorf("expression for task %d not found", id)
	}

	if expr.Len() == 1 {
		expr.Result = fmt.Sprintf("%g", value)
		expr.Status = StatusDone
		expr.Remove(el)
	} else {
		numToken := NumToken{value}
		expr.InsertBefore(numToken, el)
		expr.Remove(el)
		cs.extractTasksFromExpression(expr)
	}

	return nil
}

func (cs *CalcService) extractTasksFromExpression(expr *Expression) int {
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

		task := new(task.Task)
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
