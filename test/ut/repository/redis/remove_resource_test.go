package repository_test

import (
	"context"
	"testing"

	"github.com/dropboks/user-service/internal/domain/repository"
	"github.com/dropboks/user-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RemoveResourceRepositorySuite struct {
	suite.Suite
	redisRepository repository.RedisRepository
	mockRedisClient *mocks.MockRedisCache
}

func (s *RemoveResourceRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	redisClient := new(mocks.MockRedisCache)
	s.mockRedisClient = redisClient
	s.redisRepository = repository.NewRedisRepository(redisClient, logger)
}

func (s *RemoveResourceRepositorySuite) SetupTest() {
	s.mockRedisClient.ExpectedCalls = nil
	s.mockRedisClient.Calls = nil
}

func TestRemoveResourceRepositorySuite(t *testing.T) {
	suite.Run(t, &RemoveResourceRepositorySuite{})
}

func (s *RemoveResourceRepositorySuite) TestAuthRepository_RemoveResource_Success() {
	key := "resource-key"
	ctx := context.Background()

	s.mockRedisClient.On("Delete", mock.Anything, key).Return(nil)

	err := s.redisRepository.RemoveResource(ctx, key)

	s.NoError(err)
	s.mockRedisClient.AssertExpectations(s.T())
}
