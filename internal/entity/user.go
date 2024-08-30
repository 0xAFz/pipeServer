package entity

import (
	"time"
)

type User struct {
	ID        int64     `json:"user_id"`
	PrivateID string    `json:"private_id"`
	PubKey    string    `json:"pubkey"`
	CreatedAt time.Time `json:"created_at"`
}
