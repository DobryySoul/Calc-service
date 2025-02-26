package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/DobryySoul/Calc-service/internal/http/models/req"
	"github.com/DobryySoul/Calc-service/internal/http/models/resp"
	"github.com/DobryySoul/Calc-service/internal/service"
	"go.uber.org/zap"
)

type Middleware func(http.Handler) http.Handler

const (
	invalidValue       = "invalid value"
	invalidContentType = "invalid content type"
	invalidExpression  = "invalid expression"
	invalidId          = "invalid id"
	expressionNotFound = "expression not found"
	emptyQueue         = "no tasks in queue"
	invalidResultInput = "invalid result"
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
		expr          req.ExpressionRequest
		responseError resp.ResponseError
		answer        resp.Created
	)
	
	if !slices.Contains(r.Header["Content-Type"], "application/json") {
		w.WriteHeader(http.StatusInternalServerError)
		
		responseError.Error = invalidContentType
		
		err := json.NewEncoder(w).Encode(responseError)
		if err == nil {
			cs.log.Error("can't encode error", zap.Error(err))
		}
		return
	}
	
	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		
		responseError.Error = invalidExpression
		
		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
	
	id, err := cs.CalcService.AddExpression(expr.Expression)
	
	cs.log.Info("received expression", zap.Int("id", id), zap.String("expression", expr.Expression))
	
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		
		responseError.Error = invalidExpression
		
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

func (cs *calcStates) listAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError resp.ResponseError

	list := cs.CalcService.ListAll()
	cs.log.Info("received list of expressions", zap.Int("length", len(list.Exprs)))

	err := json.NewEncoder(w).Encode(&list)
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

	var responseError resp.ResponseError

	id := r.PathValue("id")
	Id, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = invalidId

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	expr, err := cs.CalcService.FindById(Id)
	if err != nil {
		cs.log.Error("expression not found by id", zap.Int("id", Id), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = expressionNotFound

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&expr)
	if err != nil {
		cs.log.Error("could not encode expression", zap.Int("id", Id), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}

func (cs *calcStates) sendTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var responseError resp.ResponseError

	cs.log.Info("fetching new task from queue")

	newTask := cs.CalcService.GetTask()
	if newTask == nil {
		cs.log.Warn("no tasks in queue")
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = emptyQueue

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("task fetched", zap.Int("task_id", newTask.ID))

	answer := struct {
		Task *resp.Task `json:"task"`
	}{
		Task: newTask,
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&answer)
	if err != nil {
		cs.log.Error("error encoding task response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("task sent successfully", zap.Int("task_id", newTask.ID), zap.String("task", newTask.Arg1+" "+newTask.Operation+" "+newTask.Arg2))
}

func (cs *calcStates) receiveResult(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var (
		res           req.Result
		responseError resp.ResponseError
	)

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		cs.log.Error("can't decode result", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)

		responseError.Error = invalidResultInput

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("received result", zap.Int("id", res.ID), zap.Float64("value", res.Value))

	// value := res.Value

	if err = cs.CalcService.PutResult(res.ID, res.Value); err != nil {
		cs.log.Error("can't put result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}

	cs.log.Info("result put successfully", zap.Int("id", res.ID), zap.Float64("value", res.Value))

	if err = json.NewEncoder(w).Encode(res); err != nil {
		cs.log.Error("can't encode result", zap.Int("id", res.ID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		responseError.Error = err.Error()

		_ = json.NewEncoder(w).Encode(responseError)
		return
	}
}
