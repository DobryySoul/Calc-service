package service

import (
	"container/list"
)

const (
	StatusError     = "Error"
	StatusDone      = "Done"
	StatusInProcess = "In process"
)

type Expression struct {
	*list.List
	ID     string `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

// структура для ответа на endpoint /expression/:id
type ExpressionAnswer struct {
	Expr Expression `json:"expression"`
}

// структура для ответа на endpoint /expressions
type ExpressionList struct {
	Exprs []Expression `json:"expressions"`
}

func NewExpression(id, expr string) (*Expression, error) {
	
}
