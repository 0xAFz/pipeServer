package services

import (
	"context"
	"pipe/internal/entity"
	"pipe/internal/repository"
)

type MessageService struct {
	messageRepository repository.Message
	redisRepository   repository.RedisRepository
}

func NewMessageService(messageRepository repository.Message, redisRepository repository.RedisRepository) *MessageService {
	return &MessageService{messageRepository: messageRepository, redisRepository: redisRepository}
}

func (m *MessageService) Send(message entity.Message) error {
	return m.messageRepository.Send(message)
}

func (m *MessageService) GetUserMessages(ID int64) ([]entity.Message, error) {
	return m.messageRepository.ByUserID(ID)
}

func (m *MessageService) DeleteAll(ID int64) error {
	return m.messageRepository.DeleteAllByUserID(ID)
}

func (m *MessageService) AddToRedis(ctx context.Context, userID int64, message string) error {
	return m.redisRepository.PushMessage(ctx, userID, message)
}

func (m *MessageService) GetRedisMessages(ctx context.Context, userID, start, stop int64) ([]string, error) {
	return m.redisRepository.GetMessages(ctx, userID, start, stop)
}

func (m *MessageService) ListenForNewMessage(ctx context.Context, userID int64, timeout float64) (string, error) {
	return m.redisRepository.WaitForNewMessage(ctx, userID, timeout)
}
