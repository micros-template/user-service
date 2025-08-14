package repository

import (
	"context"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/infrastructure/cache"
	"10.1.20.130/dropping/user-service/internal/infrastructure/logger"
	"github.com/rs/zerolog"
)

type (
	RedisRepository interface {
		SetResource(context.Context, string, string, time.Duration) error
	}
	redisRepository struct {
		redisClient cache.RedisCache
		logger      zerolog.Logger
		logEmitter  logger.LoggerInfra
	}
)

func NewRedisRepository(r cache.RedisCache, logEmitter logger.LoggerInfra, logger zerolog.Logger) RedisRepository {
	return &redisRepository{
		redisClient: r,
		logger:      logger,
		logEmitter:  logEmitter,
	}
}

func (a *redisRepository) SetResource(c context.Context, key, value string, duration time.Duration) error {
	err := a.redisClient.Set(c, key, value, duration)
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_SET_RESOURCE.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_SET_RESOURCE
	}
	return nil
}
