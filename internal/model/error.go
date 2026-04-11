package model

import "errors"

var (
	ErrNotFound           = errors.New("record not found")
	ErrAlreadyExists      = errors.New("record already exists")
	ErrChatTitleIsEmpty   = errors.New("chat title is empty")
	ErrChatTitleTooLong   = errors.New("chat title too long")
	ErrMessageTextIsEmpty = errors.New("message text is empty")
	ErrMessageTextTooLong = errors.New("message text too long")
)
