package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

func RPN(expression string) ([]string, error) {
	tokens := createToken(expression)
	output, err := convertingAnExpression(tokens)
	if err != nil {
		return nil, err
	}

	return output, nil
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
