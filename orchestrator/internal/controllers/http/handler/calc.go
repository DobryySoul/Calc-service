package handler

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/DobryySoul/orchestrator/internal/controllers/http/models/req"
	"github.com/DobryySoul/orchestrator/internal/controllers/http/models/resp"
	"github.com/DobryySoul/orchestrator/internal/service"
	"go.uber.org/zap"
)

type Middleware func(http.Handler) http.Handler

type calcHandlers struct {
	CalcService *service.CalcService
	log         *zap.Logger
}

func NewCalcHandler(log *zap.Logger, calcService *service.CalcService) *calcHandlers {
	return &calcHandlers{
		CalcService: calcService,
		log:         log,
	}
}

func Middlewares(next http.Handler, ds ...Middleware) http.Handler {
	middleware := next
	for d := len(ds) - 1; d >= 1; d-- {
		middleware = ds[d](middleware)
	}

	return middleware
}

func (cs *calcHandlers) Calculate(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var (
		expr          req.ExpressionRequest
		responseError resp.ResponseError
		answer        resp.Created
	)

	if !slices.Contains(r.Header["Content-Type"], "application/json") {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = invalidContentType

		err := json.NewEncoder(w).Encode(responseError)
		if err == nil {
			cs.log.Error("can't encode error", zap.Error(err))
		}
		return
	}

	err = json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = invalidExpression

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	id, err := cs.CalcService.AddExpression(expr.Expression, userID)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = err.Error() // ErrEmptyExpression

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	answer.Id = id

	cs.log.Info("sent answer", zap.Int("id", answer.Id))

	err = json.NewEncoder(w).Encode(&answer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

}

func (cs *calcHandlers) ListAll(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError resp.ResponseError

	list := cs.CalcService.ListAll(userID)
	cs.log.Info("received list of expressions", zap.Int("length", len(list.Exprs)))

	err = json.NewEncoder(w).Encode(&list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}

func (cs *calcHandlers) ListByID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError resp.ResponseError

	id := r.PathValue("id")
	ID, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = invalidId

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	expr, err := cs.CalcService.FindById(ID, userID)
	if err != nil {
		cs.log.Error("expression not found by id", zap.Int("id", ID), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = expressionNotFound

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&expr)
	if err != nil {
		cs.log.Error("could not encode expression", zap.Int("id", ID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}

func (cs *calcHandlers) SendTask(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError resp.ResponseError

	cs.log.Info("fetching new task from queue")

	newTask := cs.CalcService.GetTaskUser(userID)
	if newTask == nil {
		cs.log.Warn("no tasks in queue")
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = emptyQueue

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	answer := struct {
		Task *resp.Task `json:"task"`
	}{
		Task: newTask,
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&answer)
	if err != nil {
		cs.log.Error("error encoding task response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("task sent successfully", zap.Int("task_id", newTask.ID), zap.String("task", newTask.Arg1+" "+newTask.Operation+" "+newTask.Arg2))
}

func (cs *calcHandlers) ReceiveResult(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var (
		res           req.Result
		responseError resp.ResponseError
	)

	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		cs.log.Error("can't decode result", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = invalidResultInput

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("received result", zap.Int("id", res.ID), zap.Any("value", res.Value))

	if err = cs.CalcService.PutResultUser(res.ID, res.Value, userID); err != nil {
		cs.log.Error("can't put result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("result put successfully", zap.Int("id", res.ID), zap.Any("value", res.Value))

	if err = json.NewEncoder(w).Encode(res); err != nil {
		cs.log.Error("can't encode result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}

// Расширение функционала, добавление статистики, собственная инициатива
func (cs *calcHandlers) GetStatistics(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		cs.log.Warn("could not find user id")
		return
	}

	_, err = strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		cs.log.Warn("could not convert string to int0", zap.String("value", cookie.Value))
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	stats := resp.Statistics{
		Operations: map[string]int{
			"+": cs.CalcService.GetOperationCount("+"),
			"-": cs.CalcService.GetOperationCount("-"),
			"*": cs.CalcService.GetOperationCount("*"),
			"/": cs.CalcService.GetOperationCount("/"),
		},
	}

	_ = json.NewEncoder(w).Encode(stats)
}
