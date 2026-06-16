package leaderboard

type OptionScore struct {
	OptionID  string `json:"option_id"`
	VoteCount int64  `json:"vote_count"`
	OrderNum  int    `json:"-"`
}

type LeaderboardResponse struct {
	RoomID string        `json:"room_id"`
	Scores []OptionScore `json:"scores"`
	Total  int64         `json:"total"`
}
