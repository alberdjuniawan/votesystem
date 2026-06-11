package option

import "errors"

var (
	ErrOptionNotFound = errors.New("option not found")
	ErrRoomNotFound   = errors.New("room not found")
	ErrNotRoomOwner   = errors.New("you are not the owner of this room")
	ErrRoomNotDraft   = errors.New("options can only be modified when room is in draft status")
	ErrInternal       = errors.New("internal server error")
)
