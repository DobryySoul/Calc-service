package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/DobryySoul/Calc-service/internal/http/models"
	"github.com/DobryySoul/Calc-service/internal/result"
	"github.com/DobryySoul/Calc-service/internal/service"
	"github.com/DobryySoul/Calc-service/internal/task"
)

type Decorator func(http.Handler) http.Handler

type calcStates struct {
	CalcService *service.CalcService
}

func NewHandler(ctx context.Context, calcService *service.CalcService) (http.Handler, error) {
	mux := http.NewServeMux()

	calcState := calcStates{CalcService: calcService}

	mux.HandleFunc("POST /api/v1/calculate", calcState.calculate)      // POST
	mux.HandleFunc("GET /api/v1/expressions", calcState.listAll)       // GET
	mux.HandleFunc("GET /api/v1/expressions/{id}", calcState.listByID) // GET
	// mux.HandleFunc("/internal/task", calcState.sendTask)
	// mux.HandleFunc("/internal/task", calcState.receiveResult)

	return mux, nil
}

func Decorate(next http.Handler, ds ...Decorator) http.Handler {
	decorated := next
	for d := len(ds) - 1; d >= 0; d-- {
		decorated = ds[d](decorated)
	}

	return decorated
}

func (calcStates *calcStates) listAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	lst := calcStates.CalcService.ListAll()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(&lst)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (calcStates *calcStates) calculate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	if !slices.Contains(r.Header["Content-Type"], "application/json") {
		http.Error(w, "Incorrect header", http.StatusUnprocessableEntity)
		return
	}

	var expr models.Expression

	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = calcStates.CalcService.AddExpression(expr.Id, expr.Expression); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var answer models.Answer
	answer.Id = expr.Id

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(&answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (cs *calcStates) listByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	expr, err := cs.CalcService.FindById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(&expr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cs *calcStates) sendTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	newTask := cs.CalcService.GetTask()
	if newTask == nil {
		http.Error(w, "no tasks", http.StatusNotFound)
		return
	}

	answer := struct {
		Task *task.Task `json:"task"`
	}{
		Task: newTask,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(&answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cs *calcStates) receiveResult(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var res result.Result
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	value, err := strconv.ParseFloat(res.Value, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	if err = cs.CalcService.PutResult(res.ID, value); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
