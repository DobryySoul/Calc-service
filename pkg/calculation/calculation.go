package calculation

import (
	// Может использоваться в других частях вашего пакета

	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Ошибки
// var (
// 	ErrEmptyExpression       = errors.New("empty expression")
// 	ErrInvalidCharacter      = errors.New("invalid character in expression")
// 	ErrMismatchedParentheses = errors.New("mismatched parentheses")
// 	ErrUnknownOperator       = errors.New("unknown operator")
// 	ErrInvalidNumber         = errors.New("invalid number format")
// 	ErrNotEnoughOperands     = errors.New("not enough operands for operator")
// 	ErrInvalidExpression     = errors.New("invalid expression")
// 	ErrDivisionByZero        = errors.New("division by zero")
// 	// Добавьте другие ошибки, если необходимо
// )

// Приоритет операторов (от высшего к низшему) - Должен совпадать с opOrder в сервисе
var operatorPriority = map[string]int{
	"+": 1, "-": 1,
	"*": 2, "/": 2,
	// "^": 3, // Если есть степень
}

// Вспомогательная функция для проверки, является ли символ оператором
func isValidOperator(ch rune) bool {
	operators := "+-*/" // Добавьте '^' если есть
	return strings.ContainsRune(operators, ch)
}

// Вспомогательная функция для валидации числа (просто проверяет формат)
func validateNumber(number string) error {
	dotCount := 0
	for _, ch := range number {
		if ch == '.' {
			dotCount++
			if dotCount > 1 {
				return ErrInvalidNumber
			}
		} else if !unicode.IsDigit(ch) {
			return ErrInvalidNumber // Может быть минус в начале, но эта функция его не обрабатывает
		}
	}
	return nil
}

// createToken - преобразует строку выражения в срез токенов (инфиксная запись).
// Учитывает унарный минус.
func createToken(expression string) ([]string, error) {
	var tokens []string
	var number strings.Builder
	expression = strings.ReplaceAll(expression, " ", "") // Удаляем пробелы для упрощения токенизации

	for _, ch := range expression {
		if unicode.IsDigit(ch) || ch == '.' {
			number.WriteRune(ch)
		} else {
			if number.Len() > 0 {
				if err := validateNumber(number.String()); err != nil {
					return nil, fmt.Errorf("%w: %s", err, number.String())
				}
				tokens = append(tokens, number.String())
				number.Reset()
			}
			if !unicode.IsSpace(ch) { // Пробелы уже удалены, но проверка не помешает
				token := string(ch)

				// Обработка унарного минуса:
				// Минус считается унарным, если он:
				// - В начале выражения
				// - После открывающей скобки '('
				// // - После оператора (кроме самого себя)
				// isUnaryMinus := token == "-" && (i == 0 ||
				// 	(i > 0 && (expression[i-1] == '(' || isValidOperator(rune(expression[i-1])))))

				if isValidOperator(ch) || ch == '(' || ch == ')' {
					tokens = append(tokens, token)
				} else {
					return nil, fmt.Errorf("%w: %c", ErrInvalidCharacter, ch)
				}
			}
		}
	}

	if number.Len() > 0 {
		if err := validateNumber(number.String()); err != nil {
			return nil, fmt.Errorf("%w: %s", err, number.String())
		}
		tokens = append(tokens, number.String())
	}

	// TODO: Отдельная обработка унарного минуса после токенизации
	// Текущая простая токенизация может неверно обработать унарный минус.
	// Более надежно: после токенизации пройтись по токенам и пометить/изменить
	// токены минуса, которые являются унарными (например, преобразовать "-5" в "-5" токен,
	// или в токены "0", "-", "5").
	// Для простоты текущего примера, предполагаем, что на входе нет сложных унарных минусов типа "5*-2".

	return tokens, nil
}

// convertingAnExpression - преобразует токены из инфиксной записи в токены ОПЗ.
func convertingAnExpression(tokens []string) ([]string, error) {
	var output []string    // Выходная очередь (будет слайсом)
	var operators []string // Стек операторов

	// Используем определенный выше operatorPriority

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			// Если токен - число, добавляем его в выходную очередь
			output = append(output, token)
		} else if token == "(" {
			// Если токен - открывающая скобка, помещаем ее в стек операторов
			operators = append(operators, token)
		} else if token == ")" {
			// Если токен - закрывающая скобка
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				// Перемещаем операторы из стека в выходную очередь до открывающей скобки
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1] // Извлекаем из стека
			}
			if len(operators) == 0 {
				// Если открывающая скобка не найдена
				return nil, ErrMismatchedParentheses
			}
			// Удаляем открывающую скобку из стека
			operators = operators[:len(operators)-1]
		} else { // Если токен - оператор
			priority, isOperator := operatorPriority[token]
			if !isOperator {
				// Это должна быть ошибка, т.к. все символы, не являющиеся числами, скобками или операторами, должны отлавливаться в createToken
				// Но как защитный механизм:
				return nil, fmt.Errorf("%w: %s", ErrUnknownOperator, token)
			}

			for len(operators) > 0 {
				lastOp := operators[len(operators)-1]
				// Если последний оператор в стеке не является скобкой И его приоритет >= текущего
				if lastOp != "(" && operatorPriority[lastOp] >= priority {
					output = append(output, lastOp)
					operators = operators[:len(operators)-1] // Извлекаем из стека
				} else {
					break // Иначе останавливаем перемещение
				}
			}
			// Помещаем текущий оператор в стек
			operators = append(operators, token)
		}
	}

	// Перемещаем оставшиеся операторы из стека в выходную очередь
	for len(operators) > 0 {
		lastOp := operators[len(operators)-1]
		if lastOp == "(" {
			// Если остались открывающие скобки в стеке
			return nil, ErrMismatchedParentheses
		}
		output = append(output, lastOp)
		operators = operators[:len(operators)-1] // Извлекаем из стека
	}

	return output, nil // Возвращаем токены в ОПЗ
}

// evaluateRPN - вычисляет результат выражения в ОПЗ.
// Возвращает слайс строк с ОДНИМ элементом (результатом) при успехе.
func evaluateRPN(tokens []string) ([]string, error) {
	var stack []float64 // Стек для чисел при вычислении ОПЗ

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			// Если токен - число, помещаем его в стек
			stack = append(stack, num)
		} else { // Если токен - оператор
			// Для выполнения операции требуется как минимум два числа в стеке
			if len(stack) < 2 {
				return nil, ErrNotEnoughOperands
			}

			// Извлекаем два верхних числа из стека (b затем a)
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2] // Удаляем из стека

			var result float64
			switch token {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return nil, ErrDivisionByZero
				}
				result = a / b
				// case "^": // Если есть операция степени
				//     result = math.Pow(a, b) // Импортируйте "math"
			default:
				// Этот случай должен быть отловлен ранее, но как защитный механизм
				return nil, fmt.Errorf("%w: %s", ErrUnknownOperator, token)
			}
			// Помещаем результат операции обратно в стек
			stack = append(stack, result)
		}
	}

	// После обработки всех токенов в стеке должно остаться ровно одно число - конечный результат
	if len(stack) != 1 {
		return nil, ErrInvalidExpression // Выражение неверно сформировано
	}

	// Возвращаем результат в виде слайса строк из одного элемента
	return []string{strconv.FormatFloat(stack[0], 'f', -1, 64)}, nil
}

// RPN - Главная функция пакета calculation.
// Преобразует инфиксное выражение в ОПЗ и вычисляет результат.
// Возвращает слайс строк с ОДНИМ элементом (результатом) при успешном вычислении, или ошибку.
// *** ИСПРАВЛЕНА для возврата результата evaluateRPN ***
func RPN(expression string) ([]string, error) {
	if len(expression) == 0 {
		return nil, ErrEmptyExpression
	}

	// 1. Создание токенов инфиксной записи
	tokens, err := createToken(expression)
	if err != nil {
		return nil, err
	}

	// 2. Конвертация инфиксных токенов в ОПЗ
	output, err := convertingAnExpression(tokens) // output - это токены в ОПЗ
	if err != nil {
		return nil, err
	}

	// 3. Вычисление ОПЗ
	// evaluateRPN вернет []string с 1 элементом (результат) или ошибку.
	finalResult, err := evaluateRPN(output)
	if err != nil {
		// Если вычисление провалилось (деление на ноль, недостаточно операндов и т.п.)
		return nil, err // Возвращаем ошибку вычисления
	}

	// Если вычисление прошло успешно, возвращаем результат evaluateRPN
	// (который должен быть []string с одним элементом)
	return finalResult, nil
}
