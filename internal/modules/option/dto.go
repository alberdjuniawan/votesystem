package option

import (
	"encoding/json"
	"time"
)

type CreateOptionRequest struct {
	Label       string          `json:"label" validate:"required,min=1,max=255"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
	MediaID     string          `json:"media_id"`
	OrderNum    int             `json:"order_num"`
}

type UpdateOptionRequest struct {
	Label       string          `json:"label" validate:"required,min=1,max=255"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
	MediaID     *string         `json:"media_id"`
	OrderNum    int             `json:"order_num"`
}

type OptionResponse struct {
	ID          string          `json:"id"`
	RoomID      string          `json:"room_id"`
	Label       string          `json:"label"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	MediaID     string          `json:"media_id,omitempty"`
	MediaURL    string          `json:"media_url,omitempty"`
	OrderNum    int             `json:"order_num"`
	CreatedAt   time.Time       `json:"created_at"`
}
