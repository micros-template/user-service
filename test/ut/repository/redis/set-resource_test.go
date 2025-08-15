package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/micros-template/user-service/internal/domain/repository"
	mk "github.com/micros-template/user-service/test/mocks"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SetResourceRepositorySuite struct {
	suite.Suite
	redisRepository repository.RedisRepository
	mockRedisClient *mk.MockRedisCache
	logEmitter      *mk.LoggerInfraMock
}

func (s *SetResourceRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	redisClient := new(mk.MockRedisCache)
	mockLogEmitter := new(mk.LoggerInfraMock)

	s.mockRedisClient = redisClient
	s.logEmitter = mockLogEmitter
	s.redisRepository = repository.NewRedisRepository(redisClient, mockLogEmitter, logger)
}

func (s *SetResourceRepositorySuite) SetupTest() {
	s.mockRedisClient.ExpectedCalls = nil
	s.mockRedisClient.Calls = nil
}

func TestSetResourceRepositorySuite(t *testing.T) {
	suite.Run(t, &SetResourceRepositorySuite{})
}

func (s *SetResourceRepositorySuite) TestResourceRepository_SetResource_Success() {
	key := "resource-key"
	value := "resource-value"
	dur := 1 * time.Millisecond
	ctx := context.Background()
	s.mockRedisClient.On("Set", mock.Anything, key, value, dur).Return(nil)

	err := s.redisRepository.SetResource(ctx, key, value, dur)

	s.NoError(err)
	s.mockRedisClient.AssertExpectations(s.T())

	time.Sleep(time.Second)
	s.logEmitter.AssertExpectations(s.T())
}
