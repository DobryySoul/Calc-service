package application

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/DobryySoul/Calc-service/internal/middlewares/logger"
	"github.com/DobryySoul/Calc-service/pkg/calculation"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}

	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

func (a *Application) Run() error {
	for {
		log.Println("input expression")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("failed to read expred to read from console")
		}

		text = strings.TrimSpace(text)
		if text == "exit" {
			log.Println("application was successfully closed")
			return nil
		}

		result, err := calculation.Calc(text)
		if err != nil {
			log.Println("failed to calculate expression with error: ", err)
		} else {
			log.Printf("%s = %f", text, result)
		}
	}
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrServer           = errors.New("internal server error")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	log := slog.Default()

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		log.Warn("Invalid", "method", r.Method)

		w.WriteHeader(http.StatusMethodNotAllowed)
		jsonErrResp, _ := json.Marshal(&ErrorResponse{Error: ErrMethodNotAllowed.Error()})

		_, _ = w.Write(jsonErrResp)

		return
	}

	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Error("Failed to decode request", "error", err)

		w.WriteHeader(http.StatusUnprocessableEntity)
		jsonErrResp, _ := json.Marshal(&ErrorResponse{Error: calculation.ErrInvalidExpression.Error()})

		_, _ = w.Write(jsonErrResp)

		return
	}

	result, err := calculation.Calc(req.Expression)
	if err != nil {
		var statusCode int
		switch err {
		case
			calculation.ErrDivisionByZero,
			calculation.ErrInvalidExpression,
			calculation.ErrMismatchedParentheses,
			calculation.ErrUnknownOperator:
			statusCode = http.StatusUnprocessableEntity
			log.Warn("Calculation", "error", err)

			w.WriteHeader(statusCode)
			jsonErrResp, _ := json.Marshal(&ErrorResponse{Error: calculation.ErrInvalidExpression.Error()})

			_, _ = w.Write(jsonErrResp)
		default:
			statusCode = http.StatusInternalServerError
			log.Error("Internal server", "error", err)

			w.WriteHeader(statusCode)
			jsonErrResp, _ := json.Marshal(&ErrorResponse{Error: ErrServer.Error()})

			_, _ = w.Write(jsonErrResp)
		}

		return
	}

	var resp Response
	resp.Result = fmt.Sprintf("%.0f", result)

	jsonResp, err := json.Marshal(&resp)
	if err != nil {
		log.Error("Failed to marshal response", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		jsonErrResp, _ := json.Marshal(&ErrorResponse{Error: ErrServer.Error()})

		_, _ = w.Write(jsonErrResp)

		return
	}

	log.Info("Successful calculation", "result", resp.Result)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
}

func (a *Application) RunServer() error {
	calcHandler := http.HandlerFunc(CalcHandler)

	http.Handle("/api/v1/calculate", logger.LoggerMiddleware(slog.Default(), calcHandler))

	if err := http.ListenAndServe(":"+a.config.Addr, nil); err != nil {
		log.Fatal("failed to start server", err)

		os.Exit(1)
	}
	return nil
}
