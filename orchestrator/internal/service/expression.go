package service

import (
	"container/list"
	"fmt"
	"strconv"
	"strings"

	"github.com/DobryySoul/orchestrator/internal/http/models/resp"
	"github.com/DobryySoul/orchestrator/pkg/calculation"
)

const (
	StatusError   = "Error"
	StatusDone    = "Done"
	StatusWaiting = "Waiting"
)

const (
	TokenTypeNumber = iota
	TokenTypeOperation
	TokenTypeTask
)

type Token interface {
	Type() int
}

type (
	NumToken struct {
		Value float64
	}
	OpToken struct {
		Value string
	}
	TaskToken struct {
		ID int
	}
)

func (num NumToken) Type() int {
	return TokenTypeNumber
}

func (num OpToken) Type() int {
	return TokenTypeOperation
}

func (num TaskToken) Type() int {
	return TokenTypeTask
}

type ExprElement struct {
	ID  int
	Ptr *list.Element
}

func NewExpression(id int, expr string) (*resp.Expression, error) {
	rpn, err := calculation.RPN(expr)
	if err != nil {
		return &resp.Expression{
			ID:         id,
			Status:     StatusError,
			Result:     err.Error(),
			Expression: expr,
		}, err
	}

	expression := &resp.Expression{
		List:       list.New(),
		ID:         id,
		Status:     StatusError,
		Result:     "",
		Expression: expr,
	}

	if rpn == nil {
		return expression, nil
	}

	if len(rpn) == 1 {
		expression.Status = StatusDone
		expression.Result = rpn[0]
		return expression, nil
	}

	expression.Status = StatusWaiting
	for _, val := range rpn {
		if val == "" {
			continue
		}
		if strings.Contains("-+*/", val) {
			expression.List.PushBack(OpToken{val})
		} else {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("parse float error: %w", err)
			}
			expression.List.PushBack(NumToken{num})
		}
	}

	if err == calculation.ErrDivisionByZero {
		expression.Status = StatusError
		expression.Result = calculation.ErrDivisionByZero.Error()
	}

	return expression, nil
}
