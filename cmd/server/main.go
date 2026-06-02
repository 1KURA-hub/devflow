package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"devflow/internal/cache"
	"devflow/internal/config"
	"devflow/internal/db"
	"devflow/internal/handler"
	"devflow/internal/mq"
)

func main() {
	cfg := config.Load()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	redisClient, err := cache.OpenRedis(rootCtx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("open redis: %v", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	broker, err := mq.Open(rootCtx, cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("open rabbitmq: %v", err)
	}
	if broker != nil {
		defer broker.Close()
	}

	app := handler.NewApp(handler.Dependencies{
		Config:      cfg,
		DB:          database,
		RedisClient: redisClient,
		Broker:      broker,
		// 默认单体部署：API 进程内同时跑 worker；
		// 拆分部署时设置 DISABLE_WORKERS=true，并用 cmd/worker 单独跑消费者。
		EnableWorkers: !cfg.DisableWorkers,
		Ctx:           rootCtx,
	})
	router := handler.NewRouter(app)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Printf("devflow server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("run server: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Printf("devflow server received shutdown signal, draining...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Printf("devflow server stopped")
}
