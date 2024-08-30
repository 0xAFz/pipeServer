package entity

import "github.com/gocql/gocql"

type Message struct {
	ID       gocql.UUID `json:"message_id"`
	FromUser int64      `json:"from_user"`
	ToUser   int64      `json:"to_user"`
	Text     string     `json:"text"`
	Date     int64      `json:"date"`
}
