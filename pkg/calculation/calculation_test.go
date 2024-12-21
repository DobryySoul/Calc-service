package calculation_test

import (
	"testing"

	"github.com/DobryySoul/Calc-service/pkg/calculation"
)

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{
			name:           "Successful calculation",
			expression:     "1+1",
			expectedResult: 2,
		},
		{
			name:           "Successful calculation",
			expression:     "20+20",
			expectedResult: 40,
		},
		{
			name:           "Priority with parentheses",
			expression:     "(2+2)*2",
			expectedResult: 8,
		},
		{
			name:           "Priority",
			expression:     "2+2*2",
			expectedResult: 6,
		},
		{
			name:           "Division",
			expression:     "1/2",
			expectedResult: 0.5,
		},
		{
			name:           "Hard expression",
			expression:     "(((1/2 + 3/2) * 15 - 1) * 84) / 2 - 5 * 220",
			expectedResult: 118,
		},
	}

	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err != nil {
				t.Fatalf("successful case %s returns error", testCase.expression)
			}
			if val != testCase.expectedResult {
				t.Fatalf("%f should be equal %f", val, testCase.expectedResult)
			}
		})
	}

	testCasesFail := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:       "Simple",
			expression: "1+1*",
		},
		{
			name:       "Priority",
			expression: "2+2**2",
		},
		{
			name:       "Priority",
			expression: "((2+2-*(2",
		},
		{
			name:       "Unknown operator",
			expression: "2 + 2 ^ 2",
		},
		{
			name:       "Empty",
			expression: "",
		},
	}

	for _, testCase := range testCasesFail {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err == nil {
				t.Fatalf("expression %s is invalid but result  %f was obtained", testCase.expression, val)
			}
		})
	}
}
