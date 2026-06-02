// devflow worker 进程：单独运行 MQ 消费者，不暴露 HTTP。
//
// 与 cmd/server 共享同一份依赖装配（handler.NewApp），只是通过
// EnableWorkers=true 让消费者在本进程启动；server 进程对应地把
// EnableWorkers 设为 false 即可实现读写/异步分离部署。
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

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
		log.Fatalf("worker open database: %v", err)
	}

	redisClient, err := cache.OpenRedis(rootCtx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("worker open redis: %v", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	broker, err := mq.Open(rootCtx, cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("worker open rabbitmq: %v", err)
	}
	if broker != nil {
		defer broker.Close()
	}

	if broker == nil {
		log.Fatalf("worker requires RABBITMQ_URL to be set")
	}

	_ = handler.NewApp(handler.Dependencies{
		Config:        cfg,
		DB:            database,
		RedisClient:   redisClient,
		Broker:        broker,
		EnableWorkers: true,
		Ctx:           rootCtx,
	})

	log.Printf("devflow worker started, consuming MQ queues")
	<-rootCtx.Done()
	log.Printf("devflow worker received shutdown signal, exiting")
}
