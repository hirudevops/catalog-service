package main

import (
	"log"

	"github.com/hirudevops/catalog-service/internal/config"
	httpserver "github.com/hirudevops/catalog-service/internal/http"
)

func main() {
	cfg := config.MustLoad()

	srv, err := httpserver.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("catalog-service listening on %s", cfg.HTTPAddr)
	if err := srv.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
