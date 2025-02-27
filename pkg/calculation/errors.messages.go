package calculation

import "errors"

var (
	ErrInvalidExpression     = errors.New("expression is not valid")
	ErrDivisionByZero        = errors.New("division by zero")

	ErrMismatchedParentheses = errors.New("mismatched parentheses")
	ErrUnknownOperator       = errors.New("unknown operator")
	ErrEmptyExpression       = errors.New("expression is empty")
	ErrInvalidCharacter      = errors.New("invalid character in expression")
	ErrInvalidNumber         = errors.New("invalid number")
)
