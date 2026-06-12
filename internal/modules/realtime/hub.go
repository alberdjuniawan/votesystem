package realtime

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
)

type BroadcastMessage struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id"`
	Payload any    `json:"payload"`
}

type VoteUpdatePayload struct {
	OptionID  string `json:"option_id"`
	VoteCount int64  `json:"vote_count"`
	Total     int64  `json:"total"`
}

type client struct {
	roomID string
	send   chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*client]bool
	register   chan *client
	unregister chan *client
	broadcast  chan roomMessage
}

type roomMessage struct {
	roomID  string
	payload []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*client]bool),
		register:   make(chan *client, 256),
		unregister: make(chan *client, 256),
		broadcast:  make(chan roomMessage, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	logger.Info(ctx, "WebSocket hub started")
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "WebSocket hub stopped")
			return

		case c := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[c.roomID]; !ok {
				h.rooms[c.roomID] = make(map[*client]bool)
			}
			h.rooms[c.roomID][c] = true
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[c.roomID]; ok {
				delete(clients, c)
				close(c.send)
				if len(clients) == 0 {
					delete(h.rooms, c.roomID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.roomID]
			h.mu.RUnlock()

			for c := range clients {
				select {
				case c.send <- msg.payload:
				default:
					h.unregister <- c
				}
			}
		}
	}
}

func (h *Hub) Broadcast(roomID string, msg BroadcastMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.broadcast <- roomMessage{
		roomID:  roomID,
		payload: payload,
	}
	return nil
}

func (h *Hub) ClientCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}
