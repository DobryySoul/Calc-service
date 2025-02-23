package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/DobryySoul/Calc-service/internal/http/models"
	"github.com/DobryySoul/Calc-service/internal/service"
	"go.uber.org/zap"
)

type Middleware func(http.Handler) http.Handler

const (
	invalidValue = "invalid value"
)

type calcStates struct {
	CalcService *service.CalcService
	log         *zap.Logger
}

func NewHandler(ctx context.Context, log *zap.Logger, calcService *service.CalcService) (http.Handler, error) {
	mux := http.NewServeMux()

	calcState := calcStates{
		CalcService: calcService,
		log:         log,
	}

	mux.HandleFunc("POST /api/v1/calculate", calcState.calculate)
	mux.HandleFunc("GET /api/v1/expressions", calcState.listAll)
	mux.HandleFunc("GET /api/v1/expressions/{id}", calcState.listByID)
	mux.HandleFunc("GET /internal/task", calcState.sendTask)
	mux.HandleFunc("POST /internal/task", calcState.receiveResult)

	return mux, nil
}

func Middlewares(next http.Handler, ds ...Middleware) http.Handler {
	decorated := next
	for d := len(ds) - 1; d >= 0; d-- {
		decorated = ds[d](decorated)
	}

	return decorated
}

func (cs *calcStates) calculate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var (
		expr          models.Expression
		responseError models.ResponseError
		answer        models.Created
	)

	if !slices.Contains(r.Header["Content-Type"], "application/json") {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = "invalid content type"

		err := json.NewEncoder(w).Encode(responseError)
		if err == nil {
			cs.log.Error("could not encode error", zap.Error(err))
		}
		return
	}

	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	id, err := cs.CalcService.AddExpression(expr.Expression)

	cs.log.Info("received expression", zap.Int("id", id), zap.String("expression", expr.Expression))

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = "could not add expression"

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	answer.Id = id

	cs.log.Info("sent answer", zap.Int("id", answer.Id))

	err = json.NewEncoder(w).Encode(&answer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (cs *calcStates) listAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError models.ResponseError

	lst := cs.CalcService.ListAll()
	cs.log.Info("received list of expressions", zap.Int("length", len(lst.Exprs)))

	err := json.NewEncoder(w).Encode(&lst)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}

func (cs *calcStates) listByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError models.ResponseError

	id := r.PathValue("id")
	Id, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
	expr, err := cs.CalcService.FindById(Id)
	if err != nil {
		cs.log.Error("could not find expression by id", zap.Int("id", Id), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = err.Error()
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&expr)
	if err != nil {
		cs.log.Error("could not encode expression", zap.Int("id", Id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cs *calcStates) sendTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError models.ResponseError

	cs.log.Info("Fetching new task from queue")

	newTask := cs.CalcService.GetTask()
	if newTask == nil {
		cs.log.Warn("No tasks in queue")
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = "No tasks in queue"
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("Task fetched", zap.Int("task_id", newTask.ID))

	answer := struct {
		Task *models.Task `json:"task"`
	}{
		Task: newTask,
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&answer)
	if err != nil {
		cs.log.Error("Error encoding task response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("Task sent successfully", zap.Int("task_id", newTask.ID))
}

func (cs *calcStates) receiveResult(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var (
		res           models.Result
		responseError models.ResponseError
	)

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		cs.log.Error("could not decode result", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = err.Error()
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("received result", zap.Int("id", res.ID), zap.String("value", res.Value))

	value, err := strconv.ParseFloat(res.Value, 64)
	if err != nil {
		cs.log.Error("could not parse value", zap.String("value", res.Value), zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = invalidValue
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	if err = cs.CalcService.PutResult(res.ID, value); err != nil {
		cs.log.Error("could not put result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = "could not put result"
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	if err = json.NewEncoder(w).Encode(res); err != nil {
		cs.log.Error("could not encode result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = "could not encode result"
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}
