package repository

import (
	"context"
	"fmt"

	"github.com/redis/rueidis"
)

var _ RedisRepository = &RedisRepo{}

type RedisRepo struct {
	client rueidis.Client
}

func NewRedisRepository(redisClient rueidis.Client) RedisRepository {
	return &RedisRepo{client: redisClient}
}

func (r *RedisRepo) PushMessage(ctx context.Context, userID int64, message string) error {
	listKey := fmt.Sprintf("messages:%d", userID)
	cmd := r.client.B().Rpush().Key(listKey).Element(message).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *RedisRepo) GetMessages(ctx context.Context, userID, start, stop int64) ([]string, error) {
	listKey := fmt.Sprintf("messages:%d", userID)
	cmd := r.client.B().Lrange().Key(listKey).Start(start).Stop(stop).Build()
	messages, err := r.client.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		return nil, err
	}

	cmd = r.client.B().Ltrim().Key(listKey).Start(start).Stop(stop).Build()
	if err := r.client.Do(ctx, cmd).Error(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *RedisRepo) WaitForNewMessage(ctx context.Context, userID int64, timeout float64) (string, error) {
	listKey := fmt.Sprintf("messages:%d", userID)
	cmd := r.client.B().Blpop().Key(listKey).Timeout(timeout).Build()
	return r.client.Do(ctx, cmd).ToString()
}
