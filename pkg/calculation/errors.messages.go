package calculation

import "errors"

var (
	ErrInvalidExpression     = errors.New("expression is not valid")
	ErrDivisionByZero        = errors.New("division by zero")
	ErrMismatchedParentheses = errors.New("mismatched parentheses")
	ErrUnknownOperator       = errors.New("unknown operator")
)
