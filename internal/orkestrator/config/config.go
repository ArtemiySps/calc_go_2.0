package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AdditionTime       time.Duration
	SubtractionTime    time.Duration
	MultiplicationTime time.Duration
	DivisionTime       time.Duration

	OrkestratorPort string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		return nil, err
	}

	additionTime, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	substractionTime, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	multiplicationTime, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	divisionTime, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))

	port := os.Getenv("PORT_ORKESTRATOR")

	cfg := &Config{
		AdditionTime:       time.Duration(additionTime) * time.Millisecond,
		SubtractionTime:    time.Duration(substractionTime) * time.Millisecond,
		MultiplicationTime: time.Duration(multiplicationTime) * time.Millisecond,
		DivisionTime:       time.Duration(divisionTime) * time.Millisecond,
		OrkestratorPort:    port,
	}

	return cfg, nil
}

// также, можно обрабатывать ошибки ввода некорректных значений для переменных
