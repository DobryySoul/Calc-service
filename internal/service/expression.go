package service

import (
	"container/list"
	"strconv"
	"strings"

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

type NumToken struct {
	Value float64
}

func (num NumToken) Type() int {
	return TokenTypeNumber
}

type OpToken struct {
	Value string
}

func (num OpToken) Type() int {
	return TokenTypeOperation
}

type TaskToken struct {
	ID int
}

func (num TaskToken) Type() int {
	return TokenTypeTask
}

type Expression struct {
	*list.List
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Result     string `json:"result"`
	Expression string `json:"expression"`
}

type ExpressionUnit struct {
	Expr Expression `json:"expression"`
}

type ExpressionList struct {
	Exprs []Expression `json:"expressions"`
}

func NewExpression(id int, expr string) (*Expression, error) {
	rpn, err := calculation.NewRPN(expr)
	if err != nil {
		expression := Expression{
			List:       list.New(),
			ID:         id,
			Status:     StatusError,
			Result:     "",
			Expression: expr,
		}
		return &expression, err
	}

	if len(rpn) == 1 {
		expression := Expression{
			List:       list.New(),
			ID:         id,
			Status:     StatusDone,
			Result:     rpn[0],
			Expression: expr,
		}
		return &expression, nil
	}

	expression := Expression{
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

type ExprElement struct {
	ID  int
	Ptr *list.Element
}
