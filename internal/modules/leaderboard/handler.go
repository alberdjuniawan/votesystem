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

// GetLeaderboard godoc
// @Summary      Get room leaderboard
// @Description  Retrieves the real-time leaderboard and vote counts for a specific voting room.
// @Tags         leaderboard
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Room ID (UUID)"
// @Success      200  {object}  response.WebResponse{data=leaderboard.LeaderboardResponse} "Leaderboard retrieved successfully"
// @Failure      500  {object}  response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /rooms/{id}/leaderboard [get]
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
