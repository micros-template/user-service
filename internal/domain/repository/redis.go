package repository

import (
	"context"
	"time"

	"10.1.20.130/dropping/log-management/pkg"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/infrastructure/cache"
	"10.1.20.130/dropping/user-service/pkg/utils"
	"github.com/rs/zerolog"
)

type (
	RedisRepository interface {
		SetResource(context.Context, string, string, time.Duration) error
	}
	redisRepository struct {
		redisClient cache.RedisCache
		logger      zerolog.Logger
		logEmitter  pkg.LogEmitter
		util        utils.UserServiceUtil
	}
)

func NewRedisRepository(r cache.RedisCache, logEmitter pkg.LogEmitter, util utils.UserServiceUtil, logger zerolog.Logger) RedisRepository {
	return &redisRepository{
		redisClient: r,
		logger:      logger,
		logEmitter:  logEmitter,
		util:        util,
	}
}

func (a *redisRepository) SetResource(c context.Context, key, value string, duration time.Duration) error {
	err := a.redisClient.Set(c, key, value, duration)
	if err != nil {
		go func() {
			if err := a.util.EmitLog(a.logEmitter, "ERR", dto.Err_INTERNAL_SET_RESOURCE.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_SET_RESOURCE
	}
	return nil
}
