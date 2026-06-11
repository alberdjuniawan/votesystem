package auth

import (
	"errors"

	"github.com/alberdjuniawan/votesystem/internal/middleware"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/response"
	"github.com/alberdjuniawan/votesystem/internal/shared/validator"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	result, err := h.service.Register(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			logger.Info(ctx, "Register attempt with existing email", "email", req.Email)
			response.NewError(c, response.ErrConflict, err.Error())
		default:
			logger.Error(ctx, "Register failed", "email", req.Email)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "User registered successfully", "email", req.Email)
	response.Created(c, result)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	result, err := h.service.Login(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			logger.Warn(ctx, "Failed login attempt", "email", req.Email)
			response.NewError(c, response.ErrUnauthorized, nil)
		default:
			logger.Error(ctx, "Login failed", "email", req.Email)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "User logged in successfully", "email", req.Email)
	response.OK(c, result)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)

	ctx := c.Request.Context()
	result, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			logger.Warn(ctx, "GetMe: valid token but user not found", "user_id", userID)
			response.NewError(c, response.ErrUnauthorized, "session expired or user not found")
		default:
			logger.Error(ctx, "GetMe failed", "user_id", userID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	response.OK(c, result)
}
