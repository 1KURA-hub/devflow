package main

import (
	"context"
	"log"

	"devflow/internal/cache"
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

	redisClient, err := cache.OpenRedis(context.Background(), cfg.RedisAddr)
	if err != nil {
		log.Fatalf("open redis: %v", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	router := handler.NewRouter(handler.Dependencies{
		Config:      cfg,
		DB:          database,
		RedisClient: redisClient,
	})

	log.Printf("devflow server listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
