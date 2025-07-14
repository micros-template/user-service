package repository

import (
	"context"
	"time"

	"github.com/dropboks/user-service/internal/domain/dto"
	"github.com/dropboks/user-service/internal/infrastructure/cache"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type (
	RedisRepository interface {
		GetResource(context.Context, string) (string, error)
		SetResource(context.Context, string, string, time.Duration) error
		RemoveResource(context.Context, string) error
	}
	redisRepository struct {
		redisClient cache.RedisCache
		logger      zerolog.Logger
	}
)

func NewRedisRepository(r cache.RedisCache, logger zerolog.Logger) RedisRepository {
	return &redisRepository{
		redisClient: r,
		logger:      logger,
	}
}

func (a *redisRepository) GetResource(c context.Context, key string) (string, error) {
	v, err := a.redisClient.Get(c, key)
	if err != nil {
		if err == redis.Nil {
			return "", dto.Err_NOTFOUND_KEY_NOTFOUND
		}
		return "", dto.Err_INTERNAL_GET_RESOURCE
	}
	return v, nil
}

func (a *redisRepository) RemoveResource(c context.Context, key string) error {
	if err := a.redisClient.Delete(c, key); err != nil {
		return dto.Err_INTERNAL_DELETE_RESOURCE
	}
	return nil
}

func (a *redisRepository) SetResource(c context.Context, key, value string, duration time.Duration) error {
	err := a.redisClient.Set(c, key, value, duration)
	if err != nil {
		return dto.Err_INTERNAL_SET_RESOURCE
	}
	return nil
}
