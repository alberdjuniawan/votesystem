package room

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, service *Service, authMw gin.HandlerFunc) {
	h := NewHandler(service)

	public := rg.Group("/rooms")
	{
		public.GET("/share/:code", h.GetRoomByShareCode)
	}

	protected := rg.Group("/rooms", authMw)
	{
		protected.POST("", h.CreateRoom)
		protected.GET("", h.ListMyRooms)
		protected.GET("/:id", h.GetRoom)
		protected.PATCH("/:id/status", h.UpdateRoomStatus)
		protected.DELETE("/:id", h.DeleteRoom)
	}
}
