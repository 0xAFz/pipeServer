package services

import (
	"pipe/internal/entity"
	"pipe/internal/repositories"
)

type MessageService struct {
	repo repositories.Message
}

func NewMessageService(repo repositories.Message) *MessageService {
	return &MessageService{repo: repo}
}

func (m *MessageService) Send(message entity.Message) error {
	return m.repo.Send(message)
}

func (m *MessageService) GetUserMessages(ID int64) ([]entity.Message, error) {
	return m.repo.ByUserID(ID)
}

func (m *MessageService) DeleteAll(ID int64) error {
	return m.repo.DeleteAllByUserID(ID)
}
