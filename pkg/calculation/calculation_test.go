package calculation

import (
	"testing"
)

func TestRPN(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expected    []string
		expectError bool
	}{
		{
			name:        "Simple addition",
			expression:  "3 + 4",
			expected:    []string{"3", "4", "+"},
			expectError: false,
		},
		{
			name:        "Simple subtraction",
			expression:  "5 - 2",
			expected:    []string{"5", "2", "-"},
			expectError: false,
		},
		{
			name:        "Simple multiplication",
			expression:  "6 * 3",
			expected:    []string{"6", "3", "*"},
			expectError: false,
		},
		{
			name:        "Simple division",
			expression:  "8 / 2",
			expected:    []string{"8", "2", "/"},
			expectError: false,
		},
		{
			name:        "Expression with parentheses",
			expression:  "( 3 + 4 ) * 2",
			expected:    []string{"3", "4", "+", "2", "*"},
			expectError: false,
		},
		{
			name:        "Expression with floating point numbers",
			expression:  "3.5 + 4.2",
			expected:    []string{"3.5", "4.2", "+"},
			expectError: false,
		},
		{
			name:        "Complex expression",
			expression:  "( 3 + 4 ) * ( 2 - 1 )",
			expected:    []string{"3", "4", "+", "2", "1", "-", "*"},
			expectError: false,
		},
		{
			name:        "Mismatched parentheses",
			expression:  "( 3 + 4 ) * ( 2 - 1",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RPN(tt.expression)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !compareSlices(result, tt.expected) {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// Вспомогательная функция для сравнения слайсов
func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}