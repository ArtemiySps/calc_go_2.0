package main

import (
	"log"
	"net/http"

	"github.com/ArtemiySps/calc_go_2.0/internal/agent/config"
	a "github.com/ArtemiySps/calc_go_2.0/internal/agent/service"
	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	logger := models.MakeLogger()

	api := a.NewAgent(http.DefaultClient, cfg, logger)

	err = api.RunServer()
	if err != nil {
		log.Fatal(err.Error())
	}
}
