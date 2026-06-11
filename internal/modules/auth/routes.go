package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, service *Service, authMw gin.HandlerFunc) {
	h := NewHandler(service)

	public := rg.Group("/auth")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
	}

	protected := rg.Group("/auth", authMw)
	{
		protected.GET("/me", h.GetMe)
	}
}
