package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/DobryySoul/orchestrator/internal/config"
	"github.com/DobryySoul/orchestrator/internal/controllers/http/models/resp"
	"github.com/DobryySoul/orchestrator/internal/timeout"
	pb "github.com/DobryySoul/orchestrator/pkg/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrchestratorServiceServer interface {
	pb.UnimplementedOrchestratorServiceServer
	GetTask(context.Context, *emptypb.Empty) (*pb.Task, error)
	SendResult(context.Context, *pb.Result) (*emptypb.Empty, error)
}

type CalcService struct {
	cfg           *config.Config
	userExprTable map[uint64]map[int]*resp.Expression
	taskID        int
	userTaskTable map[uint64]map[int]ExprElement
	userTasks     map[uint64][]*resp.Task
	timeTable     map[string]time.Duration
	timeoutsTable map[int]*timeout.Timeout
	Operations    map[string]int
	mutex         sync.RWMutex
	logger        *zap.Logger
}

func NewCalcService(cfg *config.Config, logger *zap.Logger) *CalcService {
	CS := &CalcService{
		cfg:           cfg,
		userExprTable: make(map[uint64]map[int]*resp.Expression),
		userTaskTable: make(map[uint64]map[int]ExprElement),
		userTasks:     make(map[uint64][]*resp.Task),
		timeTable:     make(map[string]time.Duration),
		timeoutsTable: make(map[int]*timeout.Timeout),
		mutex:         sync.RWMutex{},
		logger:        logger,
		Operations: map[string]int{
			"+": 0,
			"-": 0,
			"*": 0,
			"/": 0,
		},
	}

	CS.timeTable["+"] = cfg.TIME_ADDITION
	CS.timeTable["-"] = cfg.TIME_SUBTRACT
	CS.timeTable["*"] = cfg.TIME_MULTIPLY
	CS.timeTable["/"] = cfg.TIME_DIVISION

	return CS
}

func (cs *CalcService) AddExpression(expr string, userID uint64) (int, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if len(expr) == 0 {
		return 0, nil
	}

	if _, ok := cs.userExprTable[userID]; !ok {
		cs.userExprTable[userID] = make(map[int]*resp.Expression)
	}

	id := 1
	if len(cs.userExprTable[userID]) > 0 {
		maxID := 0
		for exprID := range cs.userExprTable[userID] {
			if exprID > maxID {
				maxID = exprID
			}
		}
		id = maxID + 1
	}

	operations := extractOperations(expr)

	expression, err := NewExpression(id, expr)
	expression.UserID = userID

	cs.logger.Info("adding", zap.Int("id", id), zap.String("expression", expr), zap.String("status", expression.Status))

	for _, op := range operations {
		cs.Operations[op]++
	}

	cs.userExprTable[userID][id] = expression
	if err == nil && expression.Status == StatusWaiting {
		cs.extractTasksFromExpression(expression, userID)

		return id, err
	}

	return id, nil
}

func (cs *CalcService) ListAll(userID uint64) resp.ExpressionList {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	list := resp.ExpressionList{}
	for _, expr := range cs.userExprTable[userID] {
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

func (cs *CalcService) FindById(exprID int, userID uint64) (*resp.ExpressionUnit, error) {

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	expr, found := cs.userExprTable[userID][exprID]
	if !found {
		cs.logger.Error("expression not found", zap.Int("id", exprID))
		return nil, fmt.Errorf("id %d not found", exprID)
	}

	cs.logger.Info(
		"expression found",
		zap.Int("id", exprID),
		zap.String("expression", expr.Expression),
		zap.String("status", expr.Status),
	)

	return &resp.ExpressionUnit{Expr: *expr}, nil
}

func (cs *CalcService) GetTask(ctx context.Context, _ *emptypb.Empty) (*pb.Task, error) {
	const defaultTimeout = 10 * time.Second

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for userID, tasks := range cs.userTasks {
		if len(tasks) > 0 {
			newtask := tasks[0]
			cs.userTasks[userID] = tasks[1:]

			cs.logger.Info("task retrieved",
				zap.Int("task_id", newtask.ID),
				zap.String("operation_time", newtask.OperationTime.String()),
				zap.Uint64("userID", userID))

			cs.timeoutsTable[newtask.ID] = timeout.NewTimeout(
				defaultTimeout + newtask.OperationTime,
			)

			go func(task *resp.Task, userID uint64) {
				cs.mutex.Lock()
				timeout, found := cs.timeoutsTable[task.ID]
				cs.mutex.Unlock()

				if !found {
					cs.logger.Warn("timeout not found", zap.Int("task_id", task.ID))
					return
				}

				select {
				case <-timeout.Timer.C:
					cs.handleTaskTimeout(task, userID)
				case <-timeout.Ctx.Done():
					cs.logger.Info("task completed before timeout",
						zap.Int("task_id", task.ID),
						zap.Uint64("userID", userID))
				}
			}(newtask, userID)

			return &pb.Task{
				Id:            int32(newtask.ID),
				Arg1:          newtask.Arg1,
				Arg2:          newtask.Arg2,
				Operation:     newtask.Operation,
				OperationTime: durationpb.New(newtask.OperationTime),
				UserId:        userID,
			}, nil
		}
	}

	cs.logger.Warn("no tasks available")
	return nil, status.Error(codes.NotFound, "no tasks available")
}

func (cs *CalcService) handleTaskTimeout(task *resp.Task, userID uint64) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.logger.Info("task timeout has been reached",
		zap.Int("task_id", task.ID),
		zap.Uint64("userID", userID))

	cs.userTasks[userID] = append(cs.userTasks[userID], task)
	delete(cs.timeoutsTable, task.ID)
}

func (cs *CalcService) SendResult(ctx context.Context, res *pb.Result) (*emptypb.Empty, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	taskID := int(res.Id)
	userID := res.UserId

	if timeout, found := cs.timeoutsTable[taskID]; found {
		cs.logger.Info("cancelling timeout for task", zap.Int("task_id", taskID))
		timeout.Cancel()
		delete(cs.timeoutsTable, taskID)
	}

	if _, found := cs.userTaskTable[userID][taskID]; !found {
		cs.logger.Warn("task not found", zap.Int("task_id", taskID))
		return nil, status.Errorf(codes.NotFound, "task id %d not found", taskID)
	}

	var resultValue interface{}
	switch v := res.Value.(type) {
	case *pb.Result_IntResult:
		resultValue = v.IntResult
	case *pb.Result_FloatResult:
		resultValue = v.FloatResult
	case *pb.Result_Error:
		resultValue = errors.New(v.Error)
	default:
		cs.logger.Warn("unsupported result type", zap.Any("type", res.Value))
		return nil, status.Error(codes.InvalidArgument, "unsupported result type")
	}

	el := cs.userTaskTable[userID][taskID].Ptr
	exprID := cs.userTaskTable[userID][taskID].ID

	delete(cs.userTaskTable[userID], taskID)

	expr, found := cs.userExprTable[userID][exprID]
	if !found {
		cs.logger.Warn("expression not found", zap.Int("task_id", taskID))
		return nil, status.Errorf(codes.NotFound, "expression for task %d not found", taskID)
	}

	if expr.Len() == 1 {
		expr.Result = fmt.Sprintf("%v", resultValue)
		expr.Status = StatusDone
		expr.Remove(el)
	} else {
		numToken := NumToken{Value: resultValue.(float64)}
		expr.InsertBefore(numToken, el)
		expr.Remove(el)

		cs.extractTasksFromExpression(expr, userID)
	}

	return &emptypb.Empty{}, nil
}

func (cs *CalcService) GetTaskUser(userID uint64) *resp.Task {
	const defaultTimeout = 10 * time.Second

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if tasks, ok := cs.userTasks[userID]; ok && len(tasks) > 0 {
		newtask := tasks[0]
		cs.userTasks[userID] = tasks[1:]

		cs.logger.Info("task retrieved",
			zap.Int("task_id", newtask.ID),
			zap.String("operation_time", newtask.OperationTime.String()),
			zap.Uint64("userID", userID))

		cs.timeoutsTable[newtask.ID] = timeout.NewTimeout(
			defaultTimeout + newtask.OperationTime,
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
				cs.logger.Info("task timeout has been reached",
					zap.Int("task_id", task.ID),
					zap.Uint64("userID", userID))
				cs.mutex.Lock()
				cs.userTasks[userID] = append(cs.userTasks[userID], &task)
				cs.mutex.Unlock()
			case <-timeout.Ctx.Done():
				cs.logger.Info("task completed before timeout",
					zap.Int("task_id", task.ID),
					zap.Uint64("userID", userID))
				return
			}
		}(*newtask)

		return newtask
	}

	cs.logger.Warn("no tasks available for user", zap.Uint64("userID", userID))
	return nil
}

func (cs *CalcService) PutResultUser(id int, value any, userID uint64) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	timeout, found := cs.timeoutsTable[id]
	if found {
		cs.logger.Info("cancelling timeout for task", zap.Int("task_id", id))
		timeout.Cancel()
	}

	_, found = cs.userTaskTable[userID][id]
	if !found {
		cs.logger.Warn("task id %d not found", zap.Int("task_id", id))
		return fmt.Errorf("task id %d not found", id)
	}

	el := cs.userTaskTable[userID][id].Ptr
	exprID := cs.userTaskTable[userID][id].ID

	cs.logger.Info("deleting task from task table", zap.Int("task_id", id))

	delete(cs.userTaskTable[userID], id)

	expr, found := cs.userExprTable[userID][exprID]
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

		cs.extractTasksFromExpression(expr, userID)
	}

	return nil
}

func (cs *CalcService) extractTasksFromExpression(expr *resp.Expression, userID uint64) int {
	cs.logger.Info("extracting tasks from expression", zap.Int("expr_id", expr.ID), zap.Uint64("user_id", userID))

	var taskCount int
	el := expr.Front()

	if _, ok := cs.userTaskTable[userID]; !ok {
		cs.logger.Debug("creating user task table entry", zap.Uint64("user_id", userID))
		cs.userTaskTable[userID] = make(map[int]ExprElement)
	}
	if _, ok := cs.userTasks[userID]; !ok {
		cs.logger.Debug("creating user tasks slice entry", zap.Uint64("user_id", userID))
		cs.userTasks[userID] = make([]*resp.Task, 0)
	}

	for el != nil {
		el1 := el
		if el1.Value.(Token).Type() != TokenTypeNumber {
			cs.logger.Debug("skipping non-number token", zap.Any("token", el1.Value), zap.Int("expr_id", expr.ID))
			el = el.Next()
			continue
		}

		el2 := el1.Next()
		if el2 == nil || el2.Value.(Token).Type() != TokenTypeNumber {
			cs.logger.Debug("skipping, second token is not a number or end of list", zap.Any("token", el2.Value), zap.Int("expr_id", expr.ID))
			el = el.Next()
			continue
		}

		op := el2.Next()
		if op == nil || op.Value.(Token).Type() != TokenTypeOperation {
			cs.logger.Debug("skipping, third token is not an operation or end of list", zap.Any("token", op.Value), zap.Int("expr_id", expr.ID))
			el = el.Next()
			continue
		}

		task := &resp.Task{
			ID:            cs.taskID,
			Arg1:          fmt.Sprintf("%f", el1.Value.(NumToken).Value),
			Arg2:          fmt.Sprintf("%f", el2.Value.(NumToken).Value),
			Operation:     op.Value.(OpToken).Value,
			OperationTime: cs.timeTable[op.Value.(OpToken).Value] / 1e6,
			UserID:        userID,
		}

		taskElement := expr.InsertBefore(&TaskToken{ID: task.ID}, el)

		cs.userTaskTable[userID][task.ID] = ExprElement{
			ID:     expr.ID,
			Ptr:    taskElement,
			UserID: userID,
		}

		cs.userTasks[userID] = append(cs.userTasks[userID], task)

		cs.taskID++
		taskCount++

		el = op.Next()
		expr.Remove(el1)
		expr.Remove(el2)
		expr.Remove(op)

		cs.logger.Info("new task created",
			zap.Int("task_id", task.ID),
			zap.Uint64("user_id", userID),
			zap.Int("expr_id", expr.ID),
			zap.String("operation", task.Operation))
	}

	cs.logger.Info("finished extracting tasks from expression", zap.Int("expr_id", expr.ID), zap.Uint64("user_id", userID), zap.Int("task_count", taskCount))

	return taskCount
}
func (cs *CalcService) GetOperationCount(operation string) int {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	count, exists := cs.Operations[operation]
	if !exists {
		cs.logger.Warn("unknown operation", zap.String("operation", operation))
		return 0
	}

	return count
}

func extractOperations(expression string) []string {
	operators := []string{"+", "-", "*", "/"}
	foundOperators := []string{}

	for _, char := range expression {
		for _, op := range operators {
			if string(char) == op {
				foundOperators = append(foundOperators, op)
			}
		}
	}

	return foundOperators
}
