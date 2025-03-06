package calculator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ArtemiySps/calc_go_2.0/internal/application/orkestrator"
	"github.com/joho/godotenv"
)

// запрос
type ResponseFromOrkestrator struct {
	Task orkestrator.Task `json:"task"`
}

// ответ
type CalculationResponse struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

// сама функция-калькулятор
func Calculate(ctx context.Context, a, b float64, operator string, oper_time int, resultChan chan<- float64, errorChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	//fmt.Println("че то считаю")

	time.Sleep(time.Duration(oper_time) * time.Millisecond)
	select {
	case <-ctx.Done():
		return
	default:
		switch operator {
		case "+":
			resultChan <- a + b
		case "-":
			resultChan <- a - b
		case "*":
			resultChan <- a * b
		case "/":
			if b == 0 {
				errorChan <- orkestrator.ErrDivisionByZero
			}
			resultChan <- a / b
		default:
			errorChan <- orkestrator.ErrUnexpectedSymbol
		}

	}

}

// хендлер для калькульятора. Доступен по ручке "/internal/calculator"
func CalculatorHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load("./env/env_vars.env")
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	ork_port := os.Getenv("PORT_ORKESTRATOR")

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for range ticker.C {
		// fmt.Println("----get запрос")
		task_json, err := http.Get(fmt.Sprintf("http://localhost:%s/internal/task", ork_port))
		if err != nil {
			fmt.Printf("Ошибка во время GET запроса к оркестратору")
		}
		defer task_json.Body.Close()
		if task_json.StatusCode != http.StatusOK {
			continue
		}

		// fmt.Println("получил задачу")
		var task orkestrator.Task
		err = json.NewDecoder(task_json.Body).Decode(&task)
		if err != nil {
			fmt.Printf("Error decoding JSON: %v\n", err)
			return
		}

		computing_power, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

		resultChan := make(chan float64, 1)
		errorChan := make(chan error, 1)
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

		for i := 1; i <= computing_power; i++ {
			wg.Add(1)
			go Calculate(ctx, task.Arg1, task.Arg2, task.Operation, task.Operation_time, resultChan, errorChan, &wg)
		}

		var response CalculationResponse
		select {
		case result := <-resultChan:
			cancel()

			// fmt.Println("ну вот насчитал", result)
			// fmt.Println(task.ID)
			response = CalculationResponse{Result: result}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		case err := <-errorChan:
			cancel()

			response = CalculationResponse{Error: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
		}
		wg.Wait()

		// fmt.Println("отправляю оркестратору")
		calc_response, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("JSON Marshal error:%s", err)
		}

		resp, err := http.Post("http://localhost:8081/internal/task", "application/json", bytes.NewBuffer(calc_response))
		if err != nil {
			fmt.Printf("Ошибка во время POST запроса к оркестратору (результат вычислений)")
		}
		defer resp.Body.Close()
	}
}

// Запуск сервера-калькулятора
func RunCalculator() {
	// fmt.Println("калькулятор запущен")
	http.HandleFunc("/internal/calculator", CalculatorHandler)
	http.ListenAndServe(":8080", nil)
}

// curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"2+2*3+(3-(3+4)*2-4)+2\"}"
// curl -X POST http://localhost:8080/internal/calculator
