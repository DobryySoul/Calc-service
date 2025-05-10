package client

import (
	"agent/internal/models/req"
	"agent/internal/models/resp"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestClient_GetTask(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectedTask  *resp.Task
		expectedError bool
	}{
		{
			name: "successful task retrieval",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp.Task{
					ID:            1,
					Arg1:          "10",
					Arg2:          "20",
					Operation:     "+",
					OperationTime: time.Second,
				})
			},
			expectedTask: &resp.Task{
				ID:            1,
				Arg1:          "10",
				Arg2:          "20",
				Operation:     "+",
				OperationTime: time.Second,
			},
			expectedError: false,
		},
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedTask:  nil,
			expectedError: true,
		},
		{
			name: "invalid json",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectedTask:  nil,
			expectedError: true,
		},
		{
			name: "timeout",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			expectedTask:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			client := &Client{
				Client: http.Client{},
				Host:   server.URL[7:],
				Port:   "",
				Logger: logger,
			}

			task := client.GetTask()

			if tt.expectedError {
				if task != nil {
					t.Error("Expected nil task for error case")
				}
				return
			}

			if task == nil {
				t.Error("Expected non-nil task for success case")
				return
			}
		})
	}
}

func TestClient_SendResult(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	tests := []struct {
		name           string
		result         req.Result
		serverHandler  http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful result submission",
			result: req.Result{
				ID:    1,
				Value: 30.0,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var result req.Result
				if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if result.ID != 1 || result.Value != 30.0 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "server error",
			result: req.Result{
				ID:    2,
				Value: 40.0,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name: "invalid data",
			result: req.Result{
				ID:    3,
				Value: func() {},
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			var logBuffer bytes.Buffer
			zapCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(&logBuffer),
				zapcore.DebugLevel,
			)
			tempLogger := zap.New(zapCore)

			client := &Client{
				Client: http.Client{},
				Host:   server.URL[7:],
				Port:   "",
				Logger: tempLogger,
			}

			client.SendResult(tt.result)

			if tt.expectedError {
				if !bytes.Contains(logBuffer.Bytes(), []byte("error")) {
					t.Error("Expected error log for error case")
				}
			} else {
				if bytes.Contains(logBuffer.Bytes(), []byte("error")) {
					t.Error("Unexpected error log for success case")
				}
			}
		})
	}
}

func TestClient_ConnectionErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	client := &Client{
		Client: http.Client{},
		Host:   "invalid-host",
		Port:   "9999",
		Logger: logger,
	}

	t.Run("GetTask connection error", func(t *testing.T) {
		task := client.GetTask()
		if task != nil {
			t.Error("Expected nil task for connection error")
		}
	})

	t.Run("SendResult connection error", func(t *testing.T) {
		var logBuffer bytes.Buffer
		zapCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(&logBuffer),
			zapcore.DebugLevel,
		)
		tempLogger := zap.New(zapCore)
		client.Logger = tempLogger

		client.SendResult(req.Result{ID: 1, Value: 10.0})

		if !bytes.Contains(logBuffer.Bytes(), []byte("error")) {
			t.Error("Expected error log for connection error")
		}
	})
}

func TestNewRequestErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	client := &Client{
		Client: http.Client{},
		Host:   ":::invalid:::",
		Port:   "9999",
		Logger: logger,
	}

	t.Run("GetTask invalid URL", func(t *testing.T) {
		task := client.GetTask()
		if task != nil {
			t.Error("Expected nil task for invalid URL")
		}
	})

	t.Run("SendResult invalid URL", func(t *testing.T) {
		// Для перехвата логов
		var logBuffer bytes.Buffer
		zapCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(&logBuffer),
			zapcore.DebugLevel,
		)
		tempLogger := zap.New(zapCore)
		client.Logger = tempLogger

		client.SendResult(req.Result{ID: 1, Value: 10.0})

		if !bytes.Contains(logBuffer.Bytes(), []byte("error")) {
			t.Error("Expected error log for invalid URL")
		}
	})
}

func TestRequestHeaders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{
		Client: http.Client{},
		Host:   server.URL[7:],
		Port:   "",
		Logger: logger,
	}

	t.Run("GetTask headers", func(t *testing.T) {
		task := client.GetTask()
		if task != nil {
			t.Error("Expected nil task due to header check")
		}
	})

	t.Run("SendResult headers", func(t *testing.T) {
		var logBuffer bytes.Buffer
		zapCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(&logBuffer),
			zapcore.DebugLevel,
		)
		tempLogger := zap.New(zapCore)
		client.Logger = tempLogger

		client.SendResult(req.Result{ID: 1, Value: 10.0})

		if !bytes.Contains(logBuffer.Bytes(), []byte("error")) {
			t.Error("Expected error log for header check")
		}
	})
}
