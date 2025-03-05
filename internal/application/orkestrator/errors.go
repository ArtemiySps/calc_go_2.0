package orkestrator

import (
	"errors"
)

var (
	// ошибки в математическом выражении:
	ErrDivisionByZero   = errors.New("деление на ноль")
	ErrUnexpectedSymbol = errors.New("некорректный символ")

	// ошибки таскмейкера
	ErrBadExpression = errors.New("некоректное выражение")

	// ошибка
)
