package room

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

// CreateRoom godoc
// @Summary      Create a new voting room
// @Description  Creates a new voting room with specific configurations. The caller automatically becomes the room owner.
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body room.CreateRoomRequest true "Room creation details"
// @Success      201 {object} response.WebResponse{data=room.RoomResponse} "Room successfully created"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid input or max votes"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error or share code collision"
// @Router       /rooms [post]
func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	ownerID := middleware.GetUserID(c)

	result, err := h.service.CreateRoom(ctx, ownerID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidMaxVotes):
			response.NewError(c, response.ErrBadRequest, err.Error())
		case errors.Is(err, ErrShareCodeCollision):
			logger.Error(ctx, "CreateRoom failed. Share code collision", "owner_id", ownerID)
			response.NewError(c, response.ErrInternal, nil)
		default:
			logger.Error(ctx, "CreateRoom failed", "owner_id", ownerID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Room created", "room_id", result.ID, "owner_id", ownerID)
	response.Created(c, result)
}

// GetRoom godoc
// @Summary      Get room details by ID
// @Description  Retrieves the full details of a specific voting room using its UUID.
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Success      200 {object} response.WebResponse{data=room.RoomResponse} "Room details retrieved successfully"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	result, err := h.service.GetRoomByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		default:
			logger.Error(ctx, "GetRoom failed", "room_id", id)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	response.OK(c, result)
}

// GetRoomByShareCode godoc
// @Summary      Get room details by share code
// @Description  Public endpoint to retrieve room details using its unique short share code.
// @Tags         room
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "Share Code"
// @Success      200 {object} response.WebResponse{data=room.RoomResponse} "Room details retrieved successfully"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/share/{code} [get]
func (h *Handler) GetRoomByShareCode(c *gin.Context) {
	code := c.Param("code")
	ctx := c.Request.Context()

	result, err := h.service.GetRoomByShareCode(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		default:
			logger.Error(ctx, "GetRoomByShareCode failed", "code", code)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	response.OK(c, result)
}

// ListMyRooms godoc
// @Summary      List user's rooms
// @Description  Retrieves a list of all voting rooms created by the authenticated user.
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.WebResponse{data=[]room.RoomResponse} "List of rooms retrieved successfully"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms [get]
func (h *Handler) ListMyRooms(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := middleware.GetUserID(c)

	result, err := h.service.ListMyRooms(ctx, ownerID)
	if err != nil {
		logger.Error(ctx, "ListMyRooms failed", "owner_id", ownerID)
		response.NewError(c, response.ErrInternal, nil)
		return
	}

	response.OK(c, result)
}

// UpdateRoomStatus godoc
// @Summary      Update room status
// @Description  Updates the status of a room (e.g., from draft to active). Only the room owner can perform this action.
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Param        request body room.UpdateRoomStatusRequest true "New status details"
// @Success      200 {object} response.WebResponse{data=room.RoomResponse} "Room status updated successfully"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid status transition"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the room owner"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/status [patch]
func (h *Handler) UpdateRoomStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRoomStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.NewError(c, response.ErrBadRequest, err.Error())
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.NewError(c, response.ErrBadRequest, errs)
		return
	}

	ctx := c.Request.Context()
	ownerID := middleware.GetUserID(c)

	result, err := h.service.UpdateRoomStatus(ctx, id, ownerID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotOwner):
			response.NewError(c, response.ErrForbidden, err.Error())
		case errors.Is(err, ErrInvalidStatus):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "UpdateRoomStatus failed", "room_id", id, "owner_id", ownerID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Room status updated", "room_id", id, "status", req.Status)
	response.OK(c, result)
}

// DeleteRoom godoc
// @Summary      Delete a room
// @Description  Permanently deletes a room. Only the room owner can perform this action.
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID (UUID)"
// @Success      204 "No Content"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      403 {object} response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the room owner"
// @Failure      404 {object} response.WebResponse{error=response.ErrorDetail} "Room not found"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id} [delete]
func (h *Handler) DeleteRoom(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	ownerID := middleware.GetUserID(c)

	err := h.service.DeleteRoom(ctx, id, ownerID)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoomNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotOwner):
			response.NewError(c, response.ErrForbidden, err.Error())
		default:
			logger.Error(ctx, "DeleteRoom failed", "room_id", id, "owner_id", ownerID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Room deleted", "room_id", id, "owner_id", ownerID)
	response.NoContent(c)
}
