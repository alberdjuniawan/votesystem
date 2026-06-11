package auth

import (
	"context"
	"errors"

	"github.com/alberdjuniawan/votesystem/internal/config"
	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/hash"
	jwtpkg "github.com/alberdjuniawan/votesystem/internal/shared/jwt"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *Repository
	cfg  config.JWTConfig
}

func NewService(repo *Repository, cfg config.JWTConfig) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		logger.Error(ctx, "Unexpected DB error checking email", "error", err)
		return nil, ErrInternal
	}

	passwordHash, err := hash.GeneratePassword(req.Password)
	if err != nil {
		logger.Error(ctx, "Failed to hash password", "error", err)
		return nil, ErrInternal
	}

	user, err := s.repo.CreateUser(ctx, req.Email, passwordHash, req.Name)
	if err != nil {
		logger.Error(ctx, "Failed to insert user", "error", err)
		return nil, ErrInternal
	}

	return s.buildAuthResponse(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		logger.Error(ctx, "Unexpected DB error fetching user", "error", err)
		return nil, ErrInternal
	}

	if err := hash.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildAuthResponse(ctx, user)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		logger.Error(ctx, "Unexpected DB error fetching user by id", "user_id", id, "error", err)
		return nil, ErrInternal
	}

	res := toUserResponse(user)
	return &res, nil
}

func (s *Service) buildAuthResponse(ctx context.Context, user dbsqlc.User) (*AuthResponse, error) {
	userID := utils.PgUUIDToStr(user.ID)

	accessToken, refreshToken, err := jwtpkg.GenerateTokens(
		userID,
		string(user.Role),
		s.cfg.Secret,
		s.cfg.AccessExpMins,
		s.cfg.RefreshExpHours,
	)
	if err != nil {
		logger.Error(ctx, "Failed to generate JWT tokens", "user_id", userID, "error", err)
		return nil, ErrInternal
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func toUserResponse(user dbsqlc.User) UserResponse {
	return UserResponse{
		ID:    utils.PgUUIDToStr(user.ID),
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}
}
