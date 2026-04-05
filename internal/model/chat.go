package model

import "time"

type Chat struct {
	ID        uint
	Title     string
	Messages  []Message
	CreatedAt time.Time
}
