package handler

import "errors"

const (
	invalidContentType = "invalid content type"
	invalidExpression  = "invalid expression"
	invalidId          = "invalid id"
	expressionNotFound = "expression not found"
	emptyQueue         = "no tasks in queue"
	invalidResultInput = "invalid result"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)
