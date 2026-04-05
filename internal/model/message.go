package model

import "time"

type Message struct {
	ID        uint
	ChatID    uint
	Text      string
	CreatedAt time.Time
}
