package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ArtemiySps/calc_go_2.0/internal/agent/config"
	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
	"go.uber.org/zap"
)

type Agent struct {
	Client *http.Client
	Config *config.Config
	log    *zap.Logger
}

func NewAgent(client *http.Client, cfg *config.Config, logger *zap.Logger) *Agent {
	return &Agent{
		Client: client,
		Config: cfg,
		log:    logger,
	}
}

// отправляет GET-запрос оркестратору для получения задачи
func (a *Agent) GetTask() (*models.Task, error) {
	req, err := http.NewRequest("GET", "http://localhost:"+a.Config.OrkestratorPort+"/internal/task", nil)
	if err != nil {
		return nil, err
	}

	res, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, models.ErrNoTasks
		default:
			return nil, fmt.Errorf("ошибка при GET-запросе. Статус: %d", res.StatusCode)
		}
	}

	var task *models.Task
	err = json.NewDecoder(res.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// воркер
func (a *Agent) Calculate(ctx context.Context, task *models.Task, resultChan chan<- float64, errorChan chan<- error) {
	time.Sleep(time.Duration(task.Operation_time) * time.Millisecond)

	select {
	case <-ctx.Done():
		return
	default:
		switch task.Operation {
		case "+":
			resultChan <- task.Arg1 + task.Arg2
		case "-":
			resultChan <- task.Arg1 - task.Arg2
		case "*":
			resultChan <- task.Arg1 * task.Arg2
		case "/":
			if task.Arg2 == 0 {
				errorChan <- models.ErrDivisionByZero
			}
			resultChan <- task.Arg1 / task.Arg2
		default:
			errorChan <- models.ErrUnexpectedSymbol
		}

	}
}

// отправляет POST-запрос со структурой Task, где уже заполнено поле Result/Error
func (a *Agent) PostResult(task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("POST", "http://localhost:"+a.Config.OrkestratorPort+"/internal/task", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка при отправке результата оркестратору. Статус: %d", res.StatusCode)
	}

	return nil
}

func (a *Agent) Run() error {
	ticker := time.NewTicker(a.Config.GetTaskInterval)
	defer ticker.Stop()

	var ctx context.Context
	var wg sync.WaitGroup
	resultChan := make(chan float64, 1)
	errorChan := make(chan error, 1)

	for range ticker.C {
		task, err := a.GetTask()
		if err != nil {
			switch err {
			case models.ErrNoTasks:
				continue
			default:
				return err
			}
		}

		wg.Add(a.Config.ComputingPower)
		go func() {
			defer wg.Done()

			for range a.Config.ComputingPower {
				a.Calculate(ctx, task, resultChan, errorChan)
			}
		}()

		select {
		case res := <-resultChan:
			task.Result = res
		case err := <-errorChan:
			task.Error = err.Error()
		}

		if err := a.PostResult(task); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) RunServer() error {
	a.log.Info("Server (agent) starting on port " + a.Config.AgentPort)

	http.ListenAndServe(":"+a.Config.AgentPort, nil)

	if err := a.Run(); err != nil {
		return err
	}
	return nil
}

// curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"2+2*3+(3-(3+4)*2-4)+2\"}"
// curl -X POST http://localhost:8080/internal/calculator
