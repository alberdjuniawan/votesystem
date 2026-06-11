package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/alberdjuniawan/votesystem/internal/shared/db"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	miniopkg "github.com/alberdjuniawan/votesystem/internal/shared/minio"
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

	minioClient, err := miniopkg.NewClient(cfg.MinIO)
	if err != nil {
		log.Fatalf("Failed to connect to minio: %v", err)
	}
	logger.Info(ctx, "MinIO connected")

	srv := NewServer(cfg, dbPool, redisClient, minioClient)

	go func() {
		logger.Info(ctx, "Server started", "port", cfg.App.Port)
		if err := srv.Run(":" + cfg.App.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	if tel != nil {
		if err := tel.Shutdown(shutdownCtx); err != nil {
			log.Printf("Telemetry shutdown error: %v", err)
		}
	}

	logger.Info(ctx, "Server stopped")
	os.Exit(0)
}
