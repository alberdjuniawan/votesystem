package realtime

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, hub *Hub) {
	h := NewHandler(hub)

	rg.GET("/ws/rooms/:id", h.Connect)
}
