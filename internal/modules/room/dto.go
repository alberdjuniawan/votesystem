package room

import "time"

type CreateRoomRequest struct {
	Title        string     `json:"title" validate:"required,min=3,max=255"`
	Description  string     `json:"description"`
	Type         string     `json:"type" validate:"required,oneof=single_choice multiple_choice"`
	ShowRealtime bool       `json:"show_realtime"`
	MaxVotes     int        `json:"max_votes" validate:"min=1"`
	StartsAt     *time.Time `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at"`
}

type UpdateRoomStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active closed"`
}

type RoomResponse struct {
	ID           string     `json:"id"`
	OwnerID      string     `json:"owner_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	ShowRealtime bool       `json:"show_realtime"`
	MaxVotes     int        `json:"max_votes"`
	ShareCode    string     `json:"share_code"`
	ShareURL     string     `json:"share_url"`
	StartsAt     *time.Time `json:"starts_at,omitempty"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
