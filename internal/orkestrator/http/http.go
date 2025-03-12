package http

import (
	"encoding/json"
	"net/http"

	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
	"go.uber.org/zap"
)

type Service interface {
	ExpressionOperations(expr string) (float64, error)
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
		switch err {
		case models.ErrBadExpression:
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (t *TransportHttp) RunServer() {
	t.log.Info("Server (orkestrator) starting on port " + t.port)

	http.ListenAndServe(":"+t.port, nil)
}
