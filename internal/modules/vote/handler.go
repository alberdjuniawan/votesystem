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

// CastVote godoc
// @Summary      Cast a vote
// @Description  Allows an authenticated user to cast a vote for a specific option in a room. Handles both single and multiple choice limits.
// @Tags         vote
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Param        request body vote.CastVoteRequest true "Vote details"
// @Success      201 {object} response.WebResponse{data=vote.VoteResponse} "Vote successfully cast"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid input or option does not belong to the room"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Voting is closed or max votes reached"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room or option not found"
// @Failure      409 {object} response.WebResponse{error=response.ErrorDetail} "Already voted in this room or for this option"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/votes [post]
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
		case errors.Is(err, ErrAlreadyVotedOption):
			logger.Info(ctx, "Duplicate option vote attempt", "room_id", roomID, "user_id", userID)
			response.NewError(c, response.ErrConflict, err.Error())
		case errors.Is(err, ErrMaxVotesReached):
			response.NewError(c, response.ErrForbidden, err.Error())
		default:
			logger.Error(ctx, "CastVote failed", "room_id", roomID, "user_id", userID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Vote cast", "room_id", roomID, "user_id", userID, "option_id", req.OptionID)
	response.Created(c, result)
}

// GetMyVote godoc
// @Summary      Get my vote
// @Description  Retrieves the authenticated user's voting record and status for a specific room.
// @Tags         vote
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Success      200 {object} response.WebResponse{data=vote.MyVoteResponse} "User's vote retrieved successfully"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/votes/me [get]
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
