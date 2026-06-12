package realtime

import (
	"net/http"

	"github.com/alberdjuniawan/votesystem/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) Connect(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		response.NewError(c, response.ErrBadRequest, "room_id is required")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	ServeWS(c.Request.Context(), h.hub, conn, roomID)
}
