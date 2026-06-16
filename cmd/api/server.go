package main

import (
	"context"
	"net/http"
	"time"

	miniopkg "github.com/alberdjuniawan/votesystem/internal/shared/minio"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/alberdjuniawan/votesystem/docs"

	"github.com/alberdjuniawan/votesystem/internal/config"
	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/middleware"
	"github.com/alberdjuniawan/votesystem/internal/modules/auth"
	"github.com/alberdjuniawan/votesystem/internal/modules/leaderboard"
	"github.com/alberdjuniawan/votesystem/internal/modules/media"
	"github.com/alberdjuniawan/votesystem/internal/modules/option"
	"github.com/alberdjuniawan/votesystem/internal/modules/realtime"
	"github.com/alberdjuniawan/votesystem/internal/modules/room"
	"github.com/alberdjuniawan/votesystem/internal/modules/vote"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(
	cfg *config.Config,
	dbPool *pgxpool.Pool,
	redisClient *redis.Client,
	minioClient *miniopkg.Client,
	hub *realtime.Hub,
) *Server {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(otelgin.Middleware(cfg.OTel.ServiceName))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Metrics())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.OTel.ServiceName,
		})
	})

	if cfg.App.Env != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	queries := dbsqlc.New(dbPool)
	authMw := middleware.Auth(cfg.JWT.Secret)

	api := r.Group("/api/v1")

	authRepo := auth.NewRepository(queries, dbPool)
	authService := auth.NewService(authRepo, cfg.JWT)
	auth.RegisterRoutes(api, authService, authMw)

	roomRepo := room.NewRepository(queries, dbPool)
	roomService := room.NewService(roomRepo, cfg.App.BaseURL)
	room.RegisterRoutes(api, roomService, authMw)

	optionRepo := option.NewRepository(queries, dbPool)
	optionService := option.NewService(optionRepo, cfg.App.BaseURL)
	option.RegisterRoutes(api, optionService, authMw)

	mediaRepo := media.NewRepository(queries, dbPool)
	mediaService := media.NewService(mediaRepo, minioClient)
	media.RegisterRoutes(api, mediaService, authMw)

	leaderboardRepo := leaderboard.NewRepository(queries, dbPool)
	leaderboardService := leaderboard.NewService(redisClient, leaderboardRepo)
	leaderboard.RegisterRoutes(api, leaderboardService)

	realtime.RegisterRoutes(api, hub)

	voteRepo := vote.NewRepository(queries, dbPool)
	voteService := vote.NewService(voteRepo, leaderboardService, hub, dbPool)
	vote.RegisterRoutes(api, voteService, authMw)

	return &Server{
		httpServer: &http.Server{
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *Server) Run(addr string) error {
	s.httpServer.Addr = addr
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
