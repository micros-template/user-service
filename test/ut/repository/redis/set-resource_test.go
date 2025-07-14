package repository_test

import (
	"context"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SetResourceRepositorySuite struct {
	suite.Suite
	redisRepository repository.RedisRepository
	mockRedisClient *mocks.MockRedisCache
}

func (s *SetResourceRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	redisClient := new(mocks.MockRedisCache)
	s.mockRedisClient = redisClient
	s.redisRepository = repository.NewRedisRepository(redisClient, logger)
}

func (s *SetResourceRepositorySuite) SetupTest() {
	s.mockRedisClient.ExpectedCalls = nil
	s.mockRedisClient.Calls = nil
}

func TestSetResourceRepositorySuite(t *testing.T) {
	suite.Run(t, &SetResourceRepositorySuite{})
}

func (s *SetResourceRepositorySuite) TestAuthRepository_SetResource_Success() {
	key := "resource-key"
	value := "resource-value"
	dur := 1 * time.Millisecond
	ctx := context.Background()

	s.mockRedisClient.On("Set", mock.Anything, key, value, dur).Return(nil)

	err := s.redisRepository.SetResource(ctx, key, value, dur)

	s.NoError(err)
	s.mockRedisClient.AssertExpectations(s.T())
}
