package option

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, service *Service, authMw gin.HandlerFunc) {
	h := NewHandler(service)

	public := rg.Group("/rooms/:id/options")
	{
		public.GET("", h.ListOptions)
	}

	protected := rg.Group("/rooms/:id/options", authMw)
	{
		protected.POST("", h.CreateOption)
		protected.PATCH("/:optionId", h.UpdateOption)
		protected.DELETE("/:optionId", h.DeleteOption)
	}
}
