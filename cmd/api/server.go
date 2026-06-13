package main

import (
	"context"
	"net/http"

	miniopkg "github.com/alberdjuniawan/votesystem/internal/shared/minio"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

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
	r.Use(otelgin.Middleware(cfg.OTel.ServiceName))
	r.Use(middleware.RequestID())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.OTel.ServiceName,
		})
	})

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
	voteService := vote.NewService(voteRepo, leaderboardService, hub)
	vote.RegisterRoutes(api, voteService, authMw)

	_ = redisClient
	_ = minioClient

	return &Server{
		httpServer: &http.Server{
			Handler: r,
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
