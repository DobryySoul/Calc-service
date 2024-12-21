package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedCode   int
		expectedResult string
	}{
		{
			name:           "Successful calculation",
			method:         http.MethodPost,
			body:           `{"expression": "2 + 2"}`,
			expectedCode:   http.StatusOK,
			expectedResult: `{"result":"4"}`,
		},
		{
			name:           "Successful calculation 2",
			method:         http.MethodPost,
			body:           `{"expression": "(2 + 2 * 18 / 3 - 5) * 0"}`,
			expectedCode:   http.StatusOK,
			expectedResult: `{"result":"0"}`,
		},
		{
			name:           "Successful calculation with higher priority operation",
			method:         http.MethodPost,
			body:           `{"expression": "2 + 2 * 3"}`,
			expectedCode:   http.StatusOK,
			expectedResult: `{"result":"8"}`,
		},
		{
			name:           "Successful calculation with parentheses",
			method:         http.MethodPost,
			body:           `{"expression": "(3 + 2) * 2 - 1"}`,
			expectedCode:   http.StatusOK,
			expectedResult: `{"result":"9"}`,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodGet,
			body:         ``,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodPut,
			body:         ``,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodDelete,
			body:         ``,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodPatch,
			body:         ``,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Invalid JSON",
			method:       http.MethodPost,
			body:         `{"expr": "2 + 2"}`,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "Division by zero",
			method:         http.MethodPost,
			body:           `{"expression": "1 / 0"}`,
			expectedCode:   http.StatusUnprocessableEntity,
			expectedResult: `{"error":"expression is not valid"}`,
		},
		{
			name:           "Mismatched parentheses",
			method:         http.MethodPost,
			body:           `{"expression": "2 + 2 * (2 * 3"}`,
			expectedCode:   http.StatusUnprocessableEntity,
			expectedResult: `{"error":"expression is not valid"}`,
		},
		{
			name:           "Unknown operator",
			method:         http.MethodPost,
			body:           `{"expression": "2 ^ 4"}`,
			expectedCode:   http.StatusUnprocessableEntity,
			expectedResult: `{"error":"expression is not valid"}`,
		},
		{
			name:           "Expression is not valid",
			method:         http.MethodPost,
			body:           `{"expression": "2 + 4 *"}`,
			expectedCode:   http.StatusUnprocessableEntity,
			expectedResult: `{"error":"expression is not valid"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/calculate", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			CalcHandler(w, req)

			res := w.Result()
			if res.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, res.StatusCode)
			}

			if tt.expectedResult != "" {
				var responseBody map[string]interface{}
				json.NewDecoder(res.Body).Decode(&responseBody)

				if responseBody["error"] != nil {
					errorJSON, _ := json.Marshal(responseBody)
					if string(errorJSON) != tt.expectedResult {
						t.Errorf("expected error %s, got %s", tt.expectedResult, errorJSON)
					}
				}

				if responseBody["result"] != nil {
					resultJSON, _ := json.Marshal(responseBody)
					if string(resultJSON) != tt.expectedResult {
						t.Errorf("expected result %s, got %s", tt.expectedResult, resultJSON)
					}
				}
			}
		})
	}
}