package main

import (
	"github.com/ArtemiySps/calc_go_2.0/internal/application/calculator"
	"github.com/ArtemiySps/calc_go_2.0/internal/application/orkestrator"
)

func main() {
	go calculator.RunCalculator()
	orkestrator.RunOrkestrator()
}
