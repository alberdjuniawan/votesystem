package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/alberdjuniawan/votesystem/internal/db"
	"github.com/alberdjuniawan/votesystem/internal/modules/realtime"
	dbshared "github.com/alberdjuniawan/votesystem/internal/shared/db"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/metrics"
	miniopkg "github.com/alberdjuniawan/votesystem/internal/shared/minio"
	redispkg "github.com/alberdjuniawan/votesystem/internal/shared/redis"
	"github.com/alberdjuniawan/votesystem/internal/shared/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
)

// @title           VoteSystem API
// @version         1.0
// @description     API documentation for the VoteSystem platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Alberd Juniawan Pasunda
// @contact.url    https://github.com/alberdjuniawan
// @contact.email  alberd@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg := config.Load()

	logger.Init(cfg.App.Env)

	ctx := context.Background()

	tel, err := telemetry.Init(ctx, cfg.OTel.Endpoint, cfg.OTel.ServiceName, cfg.OTel.SamplerRate)
	if err != nil {
		log.Printf("Warning: failed to init telemetry: %v", err)
	}

	dbPool, err := dbshared.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	logger.Info(ctx, "Database connected")

	if err := metrics.Init(func() (int64, int64, int64) {
		s := dbPool.Stat()
		return int64(s.AcquiredConns()), int64(s.IdleConns()), int64(s.MaxConns())
	}); err != nil {
		log.Fatalf("Failed to init metrics: %v", err)
	}
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		log.Printf("Warning: failed to init runtime metrics: %v", err)
	}

	if err := db.RunMigrations(ctx, dbPool); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	logger.Info(ctx, "Database migrations applied")

	redisClient, err := redispkg.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info(ctx, "Redis connected")

	minioClient, err := miniopkg.NewClient(cfg.MinIO)
	if err != nil {
		logger.Error(ctx, "Failed to connect to minio, continuing without it", "error", err)
		minioClient = nil
	} else {
		logger.Info(ctx, "MinIO connected")
	}

	hub := realtime.NewHub()
	hubCtx, hubCancel := context.WithCancel(ctx)
	defer hubCancel()
	go hub.Run(hubCtx)

	srv := NewServer(cfg, dbPool, redisClient, minioClient, hub)

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
