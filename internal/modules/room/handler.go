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

	if req.MaxVotes == 0 {
		req.MaxVotes = 1
	}

	ctx := c.Request.Context()
	ownerID := middleware.GetUserID(c)

	result, err := h.service.CreateRoom(ctx, ownerID, req)
	if err != nil {
		logger.Error(ctx, "CreateRoom failed", "owner_id", ownerID)
		response.NewError(c, response.ErrInternal, nil)
		return
	}

	logger.Info(ctx, "Room created", "room_id", result.ID, "owner_id", ownerID)
	response.Created(c, result)
}

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
			response.NewError(c, response.ErrBadRequest, "invalid status transition. draft→active→closed only")
		default:
			logger.Error(ctx, "UpdateRoomStatus failed", "room_id", id, "owner_id", ownerID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Room status updated", "room_id", id, "status", req.Status)
	response.OK(c, result)
}

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
