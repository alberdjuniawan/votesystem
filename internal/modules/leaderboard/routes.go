package leaderboard

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, service *Service) {
	h := NewHandler(service)

	rg.GET("/rooms/:id/leaderboard", h.GetLeaderboard)
}
