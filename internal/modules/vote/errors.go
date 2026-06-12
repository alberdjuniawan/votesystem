package vote

import "errors"

var (
	ErrAlreadyVoted    = errors.New("you have already voted in this room")
	ErrRoomNotFound    = errors.New("room not found")
	ErrOptionNotFound  = errors.New("option not found")
	ErrVotingClosed    = errors.New("voting is not open")
	ErrOptionNotInRoom = errors.New("option does not belong to this room")
	ErrInternal        = errors.New("internal server error")
)
