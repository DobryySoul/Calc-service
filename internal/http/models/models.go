package models

import "time"

type Expression struct {
	// Id         int    `json:"id"`
	Expression string `json:"expression"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type Created struct {
	Id int `json:"id"`
}

type Result struct {
	ID    int    `json:"id"`
	Value string `json:"result"`
}

type Task struct {
	ID            int           `json:"id"`
	Arg1          string        `json:"arg1"`
	Arg2          string        `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}
