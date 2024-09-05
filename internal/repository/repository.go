package repository

import (
	"context"
	"pipe/internal/entity"
)

type CommonBehaviourRepository interface {
	ByID(ID int64) (entity.User, error)
	ByPrivateID(privateID string) (entity.User, error)
}

type Account interface {
	CommonBehaviourRepository
	Save(user entity.User) error
	Delete(user entity.User) error
	SetPubKey(user entity.User) error
}

type Message interface {
	ByUserID(ID int64) ([]entity.Message, error)
	DeleteAllByUserID(ID int64) error
	Send(message entity.Message) error
}

type RedisRepository interface {
	PushMessage(ctx context.Context, userID int64, message string) error
	GetMessages(ctx context.Context, userID, start, stop int64) ([]string, error)
	WaitForNewMessage(ctx context.Context, userID int64, timeout float64) (string, error)
}
