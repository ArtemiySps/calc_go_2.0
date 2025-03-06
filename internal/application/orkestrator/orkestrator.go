package orkestrator

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// создаем новый Storage для выражений и задач
var storage = NewStorage()
var task_stack TaskStack

// функция для определения приоритета оператора
func precedence(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	default:
		return 0
	}
}

// преобразование выражения в rpn
func infixToPostfix(expression string) string {
	var stack []rune
	var output strings.Builder

	for _, char := range expression {
		if unicode.IsDigit(char) {
			output.WriteRune(char)
		} else if char == '(' {
			stack = append(stack, char)
		} else if char == ')' {
			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				output.WriteRune(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1] // Удаляем '(' из стека
		} else {
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(char) {
				output.WriteRune(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)
		}
	}

	for len(stack) > 0 {
		output.WriteRune(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output.String()
}

// создает таски для калькулятора
func TaskMaker(rpn string, expr_id string) (float64, error) {
	var stack []float64

	for _, char := range rpn {
		if num, err := strconv.Atoi(string(char)); err == nil {
			stack = append(stack, float64(num))
		} else {
			if len(stack) < 2 {
				return 0, ErrBadExpression
			}

			operand2 := stack[len(stack)-1]
			operand1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			// fmt.Println(operand1, operand2, string(char))

			task := Task{
				ID:     storage.MakeTaskID(),
				Status: "need to calculate",

				Arg1:           operand1,
				Arg2:           operand2,
				Operation:      string(char),
				Operation_time: getOperationTime(char),
			}

			wait_id := task.ID
			task_stack.AddTask(task)
			// fmt.Println(task.ID, "отправил в стек")

			var res_task Task
			for res_task.Status != "done" {
				for _, el := range task_stack.Tasks {
					if el.Status == "done" && el.ID == wait_id {
						res_task = el
						task_stack.mu.Lock()
						task_stack.Tasks = task_stack.Tasks[1:]
						task_stack.mu.Unlock()
					}
				}
			}

			// fmt.Println(res_task.ID, "решена")

			if res_task.Error != "" {
				return 0, errors.New(res_task.Error)
			}

			stack = append(stack, res_task.Result)
		}
	}

	if len(stack) != 1 {
		return 0, ErrBadExpression
	}

	return stack[0], nil
}
