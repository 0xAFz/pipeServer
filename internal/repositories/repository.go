package repositories

import (
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
