package service

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/ArtemiySps/calc_go_2.0/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
	"go.uber.org/zap"
)

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type TaskStorage struct {
	mu    sync.Mutex
	Tasks map[string]models.Task
}

type ExpressionStorage struct {
	mu          sync.Mutex
	Expressions map[string]models.Expression
}

type Orkestrator struct {
	Config *config.Config
	log    *zap.Logger

	ExpressionStorage *ExpressionStorage
	TaskStorage       *TaskStorage
}

func NewExpressionStorage() *ExpressionStorage {
	strg := make(map[string]models.Expression)
	return &ExpressionStorage{
		Expressions: strg,
	}
}

func NewTaskStorage() *TaskStorage {
	strg := make(map[string]models.Task)
	return &TaskStorage{
		Tasks: strg,
	}
}

func NewOrkestrator(cfg *config.Config, logger *zap.Logger) *Orkestrator {
	return &Orkestrator{
		Config:            cfg,
		log:               logger,
		ExpressionStorage: NewExpressionStorage(),
		TaskStorage:       NewTaskStorage(),
	}
}

func (o *Orkestrator) GetOperationTime(op rune) time.Duration {
	switch op {
	case '+':
		return o.Config.AdditionTime
	case '-':
		return o.Config.SubtractionTime
	case '*':
		return o.Config.MultiplicationTime
	case '/':
		return o.Config.DivisionTime
	default:
		return 0
	}
}

func (o *Orkestrator) AddExpressionToStorage() string {
	expression := models.Expression{
		ID:     models.MakeID(),
		Status: StatusPending,
		Result: 0,
	}

	o.ExpressionStorage.Expressions[expression.ID] = expression

	return expression.ID
}

func (o *Orkestrator) ChangeExpressionStatus(id string, res float64, ok bool) {
	o.ExpressionStorage.mu.Lock()
	defer o.ExpressionStorage.mu.Unlock()
	if ok {
		if entry, okk := o.ExpressionStorage.Expressions[id]; okk {
			entry.Status = StatusCompleted
			entry.Result = res
			o.ExpressionStorage.Expressions[id] = entry
			return
		}
	}
	if entry, okk := o.ExpressionStorage.Expressions[id]; okk {
		entry.Status = StatusFailed
	}
}

func (o *Orkestrator) AddTaskToStorage(task models.Task) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()
	o.TaskStorage.Tasks[task.ID] = task
}

func (o *Orkestrator) DeleteTaskFromStorage(id string) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()
	delete(o.TaskStorage.Tasks, id)
}

func (o *Orkestrator) WaitForResult(id string) (float64, string) {
	for {
		for k, v := range o.TaskStorage.Tasks {
			if k == id && v.Status != StatusPending {
				o.DeleteTaskFromStorage(id)
				return v.Result, v.Error
			}
		}
	}
}

func (o *Orkestrator) ExpressionOperations(expr string) (float64, error) {
	id := o.AddExpressionToStorage()
	o.log.Info(id + ": added to storage")

	expr_rpn := models.InfixToPostfix(expr)
	o.log.Info(id + ": modified to postfix")

	var stack []float64
	for _, char := range expr_rpn {
		if num, err := strconv.Atoi(string(char)); err == nil {
			stack = append(stack, float64(num))
		} else {
			if len(stack) < 2 {
				o.ChangeExpressionStatus(id, 0, false)
				return 0, models.ErrBadExpression
			}

			operand2 := stack[len(stack)-1]
			operand1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			task := models.Task{
				ID:     models.MakeID(),
				Status: StatusPending,

				Arg1:           operand1,
				Arg2:           operand2,
				Operation:      string(char),
				Operation_time: int(o.GetOperationTime(char)),
			}

			o.AddTaskToStorage(task)

			res, err := o.WaitForResult(task.ID)
			if err != "" {
				o.log.Info(id + ": status changed")
				o.ChangeExpressionStatus(id, 0, false)
				return 0, errors.New(err)
			}

			stack = append(stack, res)
		}
	}

	o.log.Info(id + ": status changed")
	if len(stack) != 1 {
		o.ChangeExpressionStatus(id, 0, false)
		return 0, models.ErrBadExpression
	}

	o.ChangeExpressionStatus(id, stack[0], true)
	return stack[0], nil
}
