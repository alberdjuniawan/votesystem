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

// CreateOption godoc
// @Summary      Create a new voting option
// @Description  Adds a new option to a specific room. The room must be in "draft" status and the caller must be the room owner.
// @Tags         option
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Param        request body option.CreateOptionRequest true "Option details"
// @Success      201 {object} response.WebResponse{data=option.OptionResponse} "Option successfully created"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid input or room is not in draft status"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the room owner"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/options [post]
func (h *Handler) CreateOption(c *gin.Context) {
	roomID := c.Param("id")

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

// ListOptions godoc
// @Summary      List room options
// @Description  Retrieves all voting options associated with a specific room.
// @Tags         option
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Room ID (UUID)"
// @Success      200 {object} response.WebResponse{data=[]option.OptionResponse} "Options retrieved successfully"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/options [get]
func (h *Handler) ListOptions(c *gin.Context) {
	roomID := c.Param("id")
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

// UpdateOption godoc
// @Summary      Update an option
// @Description  Updates the details of a specific option. The room must be in "draft" status and the caller must be the room owner.
// @Tags         option
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "Room ID (UUID)"
// @Param        optionId   path      string  true  "Option ID (UUID)"
// @Param        request body option.UpdateOptionRequest true "Option update details"
// @Success      200 {object} response.WebResponse{data=option.OptionResponse} "Option successfully updated"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid input or room is not in draft status"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the room owner"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Option not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/options/{optionId} [patch]
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

// DeleteOption godoc
// @Summary      Delete an option
// @Description  Deletes a specific option. The room must be in "draft" status and the caller must be the room owner.
// @Tags         option
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "Room ID (UUID)"
// @Param        optionId   path      string  true  "Option ID (UUID)"
// @Success      204 "No Content"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Room is not in draft status"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the room owner"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Option not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/options/{optionId} [delete]
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
