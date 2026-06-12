package media

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, service *Service, authMw gin.HandlerFunc) {
	h := NewHandler(service)

	protected := rg.Group("/media", authMw)
	{
		protected.POST("", h.Upload)
		protected.DELETE("/:id", h.Delete)
	}
}
