package realtime

import (
	"context"
	"time"

	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func ServeWS(ctx context.Context, hub *Hub, conn *websocket.Conn, roomID string) {
	c := &client{
		roomID: roomID,
		send:   make(chan []byte, 256),
	}

	hub.register <- c

	go c.writePump(ctx, conn)
	c.readPump(ctx, hub, conn)
}

func (c *client) readPump(ctx context.Context, hub *Hub, conn *websocket.Conn) {
	defer func() {
		hub.unregister <- c
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				logger.Debug(ctx, "WebSocket unexpected close", "error", err)
			}
			break
		}
	}
}

func (c *client) writePump(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-c.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
