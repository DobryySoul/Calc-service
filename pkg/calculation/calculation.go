package calculation

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

func RPN(expression string) ([]string, error) {
	if len(expression) == 0 {
		return nil, ErrEmptyExpression
	}

	tokens, err := createToken(expression)
	if err != nil {
		return nil, err
	}

	output, err := convertingAnExpression(tokens)
	if err != nil {
		return nil, err
	}

	_, err = evaluateRPN(output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func createToken(expression string) ([]string, error) {
	var tokens []string
	var number strings.Builder

	for _, ch := range expression {
		if unicode.IsDigit(ch) || ch == '.' {
			number.WriteRune(ch)
		} else {
			if number.Len() > 0 {
				if err := validateNumber(number.String()); err != nil {
					return nil, err
				}
				tokens = append(tokens, number.String())
				number.Reset()
			}
			if !unicode.IsSpace(ch) {
				if !isValidOperator(ch) && ch != '(' && ch != ')' {
					return nil, ErrInvalidCharacter
				}
				tokens = append(tokens, string(ch))
			}
		}
	}

	if number.Len() > 0 {
		if err := validateNumber(number.String()); err != nil {
			return nil, err
		}
		tokens = append(tokens, number.String())
	}

	return tokens, nil
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
			if _, ok := priority[token]; !ok {
				return nil, ErrUnknownOperator
			}
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

func validateNumber(number string) error {
	dotCount := 0
	for _, ch := range number {
		if ch == '.' {
			dotCount++
			if dotCount > 1 {
				return ErrInvalidNumber
			}
		} else if !unicode.IsDigit(ch) {
			return ErrInvalidNumber
		}
	}
	return nil
}

func isValidOperator(ch rune) bool {
	operators := "+-*/"
	return strings.ContainsRune(operators, ch)
}

func evaluateRPN(tokens []string) ([]string, error) {
	var stack []float64

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return nil, errors.New("not enough operands")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
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
					return nil, ErrDivisionByZero
				}
				stack = append(stack, a/b)
			default:
				return nil, ErrUnknownOperator
			}
		}
	}

	if len(stack) != 1 {
		return nil, ErrInvalidExpression
	}

	return []string{strconv.FormatFloat(stack[0], 'f', -1, 64)}, nil
}
