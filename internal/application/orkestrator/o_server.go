package orkestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Хендлер для оркестратора. доступен по ручке "/api/v1/calculate"
func OrkestratorHandler(w http.ResponseWriter, r *http.Request) {
	//читаем запрос
	var request struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// fmt.Println("exp: ", request.Expression)

	// создаем ID для выражения
	id := storage.MakeExprID()

	// добавляем выражение в хранилище
	storage.AddExpression(id)

	// делаем rpn
	res_rpn := infixToPostfix(request.Expression)

	//fmt.Println("res_rpn: ", res_rpn)

	//вычисляем выражение
	result, err := TaskMaker(res_rpn, id)
	if err != nil {
		response := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	//fmt.Println("result: ", result)

	// изменяем статус выражения на "calculated!"
	storage.UpdateResult(id, result)

	// формируем ответ
	response := struct {
		Result float64 `json:"result"`
	}{
		Result: result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// хендлер, выдающий таск агенту
func GiveTaskHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(task_stack.Tasks)
	if len(task_stack.Tasks) != 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task_stack.Tasks[0])
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// хендлер, принимающий результаты вычислений
func GetResultHandler(w http.ResponseWriter, r *http.Request) {
	var response struct {
		Result float64 `json:"result,omitempty"`
		Error  string  `json:"error,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}
	defer r.Body.Close()

	if response.Error != "" {
		task_stack.Tasks[0].Error = response.Error
		return
	}
	task_stack.Tasks[0].Result = response.Result
	task_stack.Tasks[0].Status = "done"
}

// хендлер для получения всех выражений
func GetAllExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions := storage.GetAllExpressions()
	response := struct {
		Exprs []Expression `json:"expressions"`
	}{
		Exprs: expressions,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// хендлер для получения выражения по ID
func GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/expression/"):]
	expr := storage.GetExpression(id)
	response := struct {
		Expr *Expression `json:"expression"`
	}{
		Expr: expr,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RunOrkestrator() {
	http.HandleFunc("/api/v1/calculate", OrkestratorHandler)
	http.HandleFunc("/api/v1/expressions", GetAllExpressionsHandler)
	http.HandleFunc("/api/v1/expression/", GetExpressionHandler)
	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GiveTaskHandler(w, r)
		case http.MethodPost:
			GetResultHandler(w, r)
		}
	})
	http.ListenAndServe(":8081", nil)
}
