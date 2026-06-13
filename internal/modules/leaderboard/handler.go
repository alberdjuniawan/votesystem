package leaderboard

import (
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	roomID := c.Param("id")
	ctx := c.Request.Context()

	result, err := h.service.GetLeaderboard(ctx, roomID)
	if err != nil {
		logger.Error(ctx, "GetLeaderboard failed", "room_id", roomID, "error", err)
		response.NewError(c, response.ErrInternal, nil)
		return
	}

	response.OK(c, result)
}
