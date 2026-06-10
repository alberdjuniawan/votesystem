package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/alberdjuniawan/votesystem/internal/shared/db"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	redispkg "github.com/alberdjuniawan/votesystem/internal/shared/redis"
	"github.com/alberdjuniawan/votesystem/internal/shared/telemetry"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.App.Env)

	ctx := context.Background()

	tel, err := telemetry.Init(ctx, cfg.OTel.Endpoint, cfg.OTel.ServiceName)
	if err != nil {
		log.Printf("Warning: failed to init telemetry: %v", err)
	}

	dbPool, err := db.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	logger.Info(ctx, "Database connected")

	redisClient, err := redispkg.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info(ctx, "Redis connected")

	logger.Info(ctx, "All systems ready", "port", cfg.App.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down...")

	if tel != nil {
		if err := tel.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}
}
