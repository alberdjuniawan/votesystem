package vote

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, service *Service, authMw gin.HandlerFunc) {
	h := NewHandler(service)

	protected := rg.Group("/rooms/:id/votes", authMw)
	{
		protected.POST("", h.CastVote)
		protected.GET("/me", h.GetMyVote)
	}
}
