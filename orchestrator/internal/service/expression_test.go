package service

import (
	"testing"

	"github.com/DobryySoul/orchestrator/internal/http/models/resp"
	"github.com/DobryySoul/orchestrator/pkg/calculation"
)

func TestNewExpression(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		expr       string
		wantExpr   *resp.Expression
		wantErr    bool
		errMessage string
	}{
		{
			name: "invalid expression without second number",
			id:   1,
			expr: "2 + ",
			wantExpr: &resp.Expression{
				ID:         1,
				Status:     StatusError,
				Result:     calculation.ErrNotEnoughOperands.Error(),
				Expression: "2 + ",
			},
			wantErr: true,
		},
		{
			name: "invalid expression division by zero",
			id:   2,
			expr: "2 / 0",
			wantExpr: &resp.Expression{
				ID:         2,
				Status:     StatusError,
				Result:     calculation.ErrDivisionByZero.Error(),
				Expression: "2 / 0",
			},
			wantErr: true,
		},
		{
			name: "invalid expression without numbers",
			id:   3,
			expr: "()",
			wantExpr: &resp.Expression{
				ID:         3,
				Status:     StatusError,
				Result:     calculation.ErrInvalidExpression.Error(),
				Expression: "()",
			},
			wantErr: true,
		},
		{
			name: "invalid expression invalid character",
			id:   4,
			expr: "a - 2",
			wantExpr: &resp.Expression{
				ID:         4,
				Status:     StatusError,
				Result:     calculation.ErrInvalidCharacter.Error(),
				Expression: "a - 2",
			},
			wantErr: true,
		},
		{
			name: "invalid expression mismatched parentheses",
			id:   5,
			expr: "2 * (2 + 2",
			wantExpr: &resp.Expression{
				ID:         5,
				Status:     StatusError,
				Result:     calculation.ErrMismatchedParentheses.Error(),
				Expression: "2 * (2 + 2",
			},
			wantErr: true,
		},
		{
			name: "invalid expression unknown operator",
			id:   6,
			expr: "67 . 21",
			wantExpr: &resp.Expression{
				ID:         6,
				Status:     StatusError,
				Result:     calculation.ErrUnknownOperator.Error(),
				Expression: "67 . 21",
			},
			wantErr: true,
		},
		{
			name: "valid expression",
			id:   7,
			expr: "2 + 2",
			wantExpr: &resp.Expression{
				ID:         7,
				Status:     StatusWaiting,
				Result:     "",
				Expression: "2 + 2",
			},
			wantErr: false,
		},
		{
			name: "valid hard expression",
			id:   8,
			expr: "(3 + 5) * (12 - (4 / 2)) + (6 - (2 * 3))",
			wantExpr: &resp.Expression{
				ID:         8,
				Status:     StatusWaiting,
				Result:     "",
				Expression: "(3 + 5) * (12 - (4 / 2)) + (6 - (2 * 3))",
			},
			wantErr: false,
		},
		{
			name: "valid expression with parentheses",
			id:   9,
			expr: "(1 - 1) * 5 / 5 * 15 * 34 - 1",
			wantExpr: &resp.Expression{
				ID:         9,
				Status:     StatusWaiting,
				Result:     "",
				Expression: "(1 - 1) * 5 / 5 * 15 * 34 - 1",
			},
			wantErr: false,
		},
		{
			name: "valid expression with many parentheses",
			id:   10,
			expr: "((((((((((((((((((((1 - 1))))))))))))))))))))",
			wantExpr: &resp.Expression{
				ID:         10,
				Status:     StatusWaiting,
				Result:     "",
				Expression: "((((((((((((((((((((1 - 1))))))))))))))))))))",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExpr, err := NewExpression(tt.id, tt.expr)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMessage != "" && err.Error() != tt.errMessage {
				t.Errorf("NewExpression() error message = %v, want %v", err.Error(), tt.errMessage)
			}

			if gotExpr.ID != tt.wantExpr.ID {
				t.Errorf("NewExpression() ID = %v, want %v", gotExpr.ID, tt.wantExpr.ID)
			}

			if gotExpr.Status != tt.wantExpr.Status {
				t.Errorf("NewExpression() Status = %v, want %v", gotExpr.Status, tt.wantExpr.Status)
			}

			if gotExpr.Expression != tt.wantExpr.Expression {
				t.Errorf("NewExpression() Expression = %v, want %v", gotExpr.Expression, tt.wantExpr.Expression)
			}
		})
	}
}
