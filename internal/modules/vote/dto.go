package vote

import "time"

type CastVoteRequest struct {
	OptionID string `json:"option_id" validate:"required,uuid"`
}

type VoteResponse struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	OptionID  string    `json:"option_id"`
	CreatedAt time.Time `json:"created_at"`
}
