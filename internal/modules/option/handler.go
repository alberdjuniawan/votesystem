package option

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

func (h *Handler) CreateOption(c *gin.Context) {
	roomID := c.Param("roomId")

	var req CreateOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	result, err := h.service.CreateOption(ctx, roomID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotRoomOwner):
			response.NewError(c, response.ErrForbidden, err.Error())
		case errors.Is(err, ErrRoomNotDraft):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "CreateOption failed", "room_id", roomID, "user_id", userID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Option created", "option_id", result.ID, "room_id", roomID)
	response.Created(c, result)
}

func (h *Handler) ListOptions(c *gin.Context) {
	roomID := c.Param("roomId")
	ctx := c.Request.Context()

	result, err := h.service.ListOptions(ctx, roomID)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		default:
			logger.Error(ctx, "ListOptions failed", "room_id", roomID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	response.OK(c, result)
}

func (h *Handler) UpdateOption(c *gin.Context) {
	optionID := c.Param("optionId")

	var req UpdateOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	result, err := h.service.UpdateOption(ctx, optionID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOptionNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotRoomOwner):
			response.NewError(c, response.ErrForbidden, err.Error())
		case errors.Is(err, ErrRoomNotDraft):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "UpdateOption failed", "option_id", optionID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	response.OK(c, result)
}

func (h *Handler) DeleteOption(c *gin.Context) {
	optionID := c.Param("optionId")
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	err := h.service.DeleteOption(ctx, optionID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOptionNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotRoomOwner):
			response.NewError(c, response.ErrForbidden, err.Error())
		case errors.Is(err, ErrRoomNotDraft):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "DeleteOption failed", "option_id", optionID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Option deleted", "option_id", optionID)
	response.NoContent(c)
}
