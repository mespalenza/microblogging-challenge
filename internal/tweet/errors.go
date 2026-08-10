package tweet

import "errors"

var (
	ErrInvalidUserID  = errors.New("invalid user ID")
	ErrInvalidContent = errors.New("invalid content")
	ErrContentTooLong = errors.New("content too long")
)
