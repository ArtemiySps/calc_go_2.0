package models

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

// функция для преобразования выражения в постфиксную запись (обратная польская запись)
func InfixToPostfix(expression string) string {
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
			stack = stack[:len(stack)-1]
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

// функция для создания ID
func MakeID() string {
	return uuid.New().String()
}

func MakeLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
