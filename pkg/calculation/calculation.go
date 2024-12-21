package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

func Calc(expression string) (float64, error) {
	tokens := createToken(expression)
	output, err := convertingAnExpression(tokens)
	if err != nil {
		return 0, err
	}

	return Counting(output)
}

func createToken(expression string) []string {
	var tokens []string
	var number strings.Builder

	for _, ch := range expression {
		if unicode.IsDigit(ch) || ch == '.' {
			number.WriteRune(ch)
		} else {
			if number.Len() > 0 {
				tokens = append(tokens, number.String())
				number.Reset()
			}
			if !unicode.IsSpace(ch) {
				tokens = append(tokens, string(ch))
			}
		}
	}

	if number.Len() > 0 {
		tokens = append(tokens, number.String())
	}

	return tokens
}

func convertingAnExpression(tokens []string) ([]string, error) {
	var output []string
	var operators []string
	priority := map[string]int{
		"+": 1, "-": 1,
		"*": 2, "/": 2,
	}

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, ErrMismatchedParentheses
			}
			operators = operators[:len(operators)-1]
		} else {
			for len(operators) > 0 && priority[operators[len(operators)-1]] >= priority[token] {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return nil, ErrMismatchedParentheses
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

func Counting(tokens []string) (float64, error) {
	var stack []float64

	for _, token := range tokens {
		if value, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, value)
		} else {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
			}
			b, a := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, ErrDivisionByZero
				}
				stack = append(stack, a/b)
			default:
				return 0, ErrUnknownOperator
			}
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}

	return stack[0], nil
}
