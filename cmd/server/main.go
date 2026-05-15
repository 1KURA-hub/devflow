package main

import (
	"log"

	"devflow/internal/config"
	"devflow/internal/db"
	"devflow/internal/handler"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	router := handler.NewRouter(handler.Dependencies{
		Config: cfg,
		DB:     database,
	})

	log.Printf("devflow server listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
