package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const minProductionJWTSecretLength = 32

const defaultJWTSecret = "devflow-dev-secret"

type Config struct {
	AppEnv      string
	HTTPAddr    string
	MySQLDSN    string
	RedisAddr   string
	RabbitMQURL string
	AutoMigrate bool
	JWTSecret   string
	// DisableWorkers 若为 true，则 cmd/server 进程不再启动 MQ 消费者；
	// 用于将 worker 拆分到独立进程（cmd/worker）部署的场景。
	DisableWorkers bool
}

func Load() Config {
	cfg := Config{
		AppEnv:         getEnv("APP_ENV", "dev"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		MySQLDSN:       getEnv("MYSQL_DSN", "devflow:devflow@tcp(127.0.0.1:3307)/devflow?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:      getEnv("REDIS_ADDR", ""),
		RabbitMQURL:    getEnv("RABBITMQ_URL", ""),
		AutoMigrate:    getBoolEnv("AUTO_MIGRATE", false),
		JWTSecret:      getEnv("JWT_SECRET", defaultJWTSecret),
		DisableWorkers: getBoolEnv("DISABLE_WORKERS", false),
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

// Validate rejects unsafe production configuration before any server, worker,
// or maintenance command can connect to external dependencies.
func (c Config) Validate() error {
	if !strings.EqualFold(strings.TrimSpace(c.AppEnv), "prod") {
		return nil
	}

	secret := strings.TrimSpace(c.JWTSecret)
	if secret == "" {
		return fmt.Errorf("invalid production configuration: JWT_SECRET must not be empty")
	}
	if secret == defaultJWTSecret || secret == "change-me" {
		return fmt.Errorf("invalid production configuration: JWT_SECRET must not use a default value")
	}
	if len(secret) < minProductionJWTSecretLength {
		return fmt.Errorf(
			"invalid production configuration: JWT_SECRET must be at least %d bytes",
			minProductionJWTSecretLength,
		)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
