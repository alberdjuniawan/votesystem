package room

import "errors"

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrNotOwner           = errors.New("you are not the owner of this room")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrRoomNotDraft       = errors.New("room can only be edited in draft status")
	ErrShareCodeCollision = errors.New("failed to generate unique share code")
	ErrInternal           = errors.New("internal server error")
)
