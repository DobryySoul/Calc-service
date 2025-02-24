package resp

import (
	"container/list"
	"time"
)

type ResponseError struct {
	Error string `json:"error"`
}

type Created struct {
	Id int `json:"id"`
}

type Task struct {
	ID            int           `json:"id"`
	Arg1          string        `json:"arg1"`
	Arg2          string        `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
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
