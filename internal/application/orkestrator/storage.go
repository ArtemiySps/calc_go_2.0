package orkestrator

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	task_counter = 0
)

// структура для состояния выражения
type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

// структура таска
type Task struct {
	ID     string `json:"id"`
	Status string `json:"status"`

	Arg1           float64 `json:"arg1"`
	Arg2           float64 `json:"arg2"`
	Operation      string  `json:"operation"`
	Operation_time int     `json:"operation_time"`

	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

// стек для тасков
type TaskStack struct {
	mu    sync.RWMutex
	Tasks []Task
}

// хранилище выражений и тасков
type Storage struct {
	mu          sync.RWMutex
	Expressions map[string]*Expression
}

func (s *TaskStack) AddTask(t Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tasks = append(s.Tasks, t)
}

// создание нового хранилища
func NewStorage() *Storage {
	return &Storage{
		Expressions: make(map[string]*Expression),
	}
}

// добавление выражения в хранилище
func (s *Storage) AddExpression(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Expressions[id] = &Expression{
		ID:     id,
		Status: "calculating...",
		Result: 0,
	}
}

// обновление результата выражения
func (s *Storage) UpdateResult(id string, result float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if expr, exists := s.Expressions[id]; exists {
		expr.Result = result
		expr.Status = "calculated!"
	}
}

// получение всех выражений
func (s *Storage) GetAllExpressions() []Expression {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expressions := make([]Expression, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		expressions = append(expressions, *expr)
	}
	return expressions
}

// получение одного выражения по ID
func (s *Storage) GetExpression(id string) *Expression {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expr := s.Expressions[id]
	return expr
}

// создание id для выражения
func (s *Storage) MakeExprID() string {
	return "id_" + strconv.Itoa(len(s.Expressions))
}

// создание id для таска
func (s *Storage) MakeTaskID() string {
	task_counter++
	return "tid_" + strconv.Itoa(task_counter)
}

// получаем время для операции
func getOperationTime(op rune) int {
	err := godotenv.Load("./internal/env/env_vars.env")
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	switch op {
	case '+':
		timeAddition := os.Getenv("TIME_ADDITION_MS")
		time, _ := strconv.Atoi(timeAddition)
		return time
	case '-':
		timeSubtraction := os.Getenv("TIME_SUBTRACTION_MS")
		time, _ := strconv.Atoi(timeSubtraction)
		return time
	case '*':
		timeMultipl := os.Getenv("TIME_MULTIPLICATIONS_MS")
		time, _ := strconv.Atoi(timeMultipl)
		return time
	case '/':
		timeDivision := os.Getenv("TIME_DIVISIONS_MS")
		time, _ := strconv.Atoi(timeDivision)
		return time
	default:
		return 0
	}
}
