package repository_test

import (
	"context"
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GetResourceRepositorySuite struct {
	suite.Suite
	redisRepository repository.RedisRepository
	mockRedisClient *mocks.MockRedisCache
}

func (s *GetResourceRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	redisClient := new(mocks.MockRedisCache)
	s.mockRedisClient = redisClient
	s.redisRepository = repository.NewRedisRepository(redisClient, logger)
}

func (s *GetResourceRepositorySuite) SetupTest() {
	s.mockRedisClient.ExpectedCalls = nil
	s.mockRedisClient.Calls = nil
}

func TestGetResourceRepositorySuite(t *testing.T) {
	suite.Run(t, &GetResourceRepositorySuite{})
}

func (s *GetResourceRepositorySuite) TestAuthRepository_GetResource_Success() {
	key := "resource-key"
	value := "resource-value"
	ctx := context.Background()
	s.mockRedisClient.On("Get", mock.Anything, key).Return(value, nil)

	val, err := s.redisRepository.GetResource(ctx, key)

	s.NoError(err)
	s.NotEmpty(val)

	s.mockRedisClient.AssertExpectations(s.T())
}

func (s *GetResourceRepositorySuite) TestAuthRepository_GetResource_NotFound() {
	key := "resource-key"
	ctx := context.Background()

	s.mockRedisClient.On("Get", mock.Anything, key).Return("", redis.Nil)

	val, err := s.redisRepository.GetResource(ctx, key)

	s.Error(err)
	s.Empty(val)
	s.mockRedisClient.AssertExpectations(s.T())
}
