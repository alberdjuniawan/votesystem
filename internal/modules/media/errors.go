package media

import "errors"

var (
	ErrFileTooLarge    = errors.New("file size exceeds 5MB limit")
	ErrInvalidMimeType = errors.New("only image files are allowed (jpeg, png, webp)")
	ErrMediaNotFound   = errors.New("media not found")
	ErrNotUploader     = errors.New("you are not the uploader of this media")
	ErrInternal        = errors.New("internal server error")
)
