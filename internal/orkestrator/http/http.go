package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
	"go.uber.org/zap"
)

type Service interface {
	ExpressionOperations(expr string) (float64, error)
	GiveTask() (models.Task, error)
	ChangeTask(id string, task models.Task)
}

type TransportHttp struct {
	s    Service
	port string
	log  *zap.Logger
}

func NewTransportHttp(s Service, port string, logger *zap.Logger) *TransportHttp {
	t := &TransportHttp{
		s:    s,
		port: port,
		log:  logger,
	}

	http.HandleFunc("/api/v1/calculate", t.OrkestratorHandler)
	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t.GiveTaskHandler(w, r)
		case http.MethodPost:
			t.GetResultHandler(w, r)
		}
	})

	return t
}

// Хендлер для оркестратора. доступен по ручке "/api/v1/calculate"
func (t *TransportHttp) OrkestratorHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := t.s.ExpressionOperations(request.Expression)
	if err != nil {
		t.log.Error(err.Error())
		switch err {
		case models.ErrBadExpression:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case models.ErrUnexpectedSymbol:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data, err := json.Marshal(map[string]any{
		"result": res,
	})
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (t *TransportHttp) GiveTaskHandler(w http.ResponseWriter, r *http.Request) {
	t.log.Info("Ready to give task")

	task, err := t.s.GiveTask()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Println(task, "jkljkljkl")

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(task)
	// data, err := json.Marshal(struct {
	// 	Task models.Task `json:"task"`
	// }{
	// 	Task: task,
	// })
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// _, err = w.Write(data)
	// if err != nil {
	// 	t.log.Error(err.Error())
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	t.log.Info("Gave task " + task.ID)
}

func (t *TransportHttp) GetResultHandler(w http.ResponseWriter, r *http.Request) {
	t.log.Info("Ready to get result")

	defer r.Body.Close()

	var result models.Task

	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t.s.ChangeTask(result.ID, result)
	t.log.Info("Got result for task " + result.ID)
}

func (t *TransportHttp) RunServer() {
	t.log.Info("Server (orkestrator) starting on port " + t.port)

	http.ListenAndServe(":"+t.port, nil)
}
