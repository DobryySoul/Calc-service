package service

import (
	"container/list"
	"strconv"
	"strings"

	"github.com/DobryySoul/Calc-service/internal/http/models/resp"
	"github.com/DobryySoul/Calc-service/pkg/calculation"
)

const (
	StatusError   = "Error"
	StatusDone    = "Done"
	StatusPending = "Pending"
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
		expression := resp.Expression{
			List:       list.New(),
			ID:         id,
			Status:     StatusError,
			Result:     "",
			Expression: expr,
		}
		return &expression, err
	}

	if len(rpn) == 1 {
		expression := resp.Expression{
			List:       list.New(),
			ID:         id,
			Status:     StatusDone,
			Result:     rpn[0],
			Expression: expr,
		}
		return &expression, nil
	}

	expression := resp.Expression{
		List:       list.New(),
		ID:         id,
		Status:     StatusPending,
		Result:     "",
		Expression: expr,
	}
	for _, val := range rpn {
		if strings.Contains("-+*/", val) {
			expression.PushBack(OpToken{val})
		} else {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, err
			}
			expression.PushBack(NumToken{num})
		}
	}
	return &expression, nil
}
