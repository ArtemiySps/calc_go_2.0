package main

import (
	"log"

	"github.com/ArtemiySps/calc_go_2.0/internal/orkestrator/config"
	h "github.com/ArtemiySps/calc_go_2.0/internal/orkestrator/http"
	"github.com/ArtemiySps/calc_go_2.0/internal/orkestrator/service"
	"github.com/ArtemiySps/calc_go_2.0/pkg/models"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	logger := models.MakeLogger()

	api := service.NewOrkestrator(cfg, logger)

	logger = models.MakeLogger()

	transport := h.NewTransportHttp(api, cfg.OrkestratorPort, logger)

	transport.RunServer()
}
