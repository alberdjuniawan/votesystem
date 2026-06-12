package vote

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

func (h *Handler) CastVote(c *gin.Context) {
	roomID := c.Param("id")

	var req CastVoteRequest
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

	result, err := h.service.CastVote(ctx, roomID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyVoted):
			logger.Info(ctx, "Duplicate vote attempt", "room_id", roomID, "user_id", userID)
			response.NewError(c, response.ErrConflict, err.Error())
		case errors.Is(err, ErrVotingClosed):
			response.NewError(c, response.ErrForbidden, err.Error())
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrOptionNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrOptionNotInRoom):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "CastVote failed", "room_id", roomID, "user_id", userID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Vote cast", "room_id", roomID, "user_id", userID, "option_id", req.OptionID)
	response.Created(c, result)
}

func (h *Handler) GetMyVote(c *gin.Context) {
	roomID := c.Param("id")
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	result, err := h.service.GetMyVote(ctx, roomID, userID)
	if err != nil {
		logger.Error(ctx, "GetMyVote failed", "room_id", roomID, "user_id", userID)
		response.NewError(c, response.ErrInternal, nil)
		return
	}

	if result == nil {
		response.OK(c, nil)
		return
	}

	response.OK(c, result)
}
