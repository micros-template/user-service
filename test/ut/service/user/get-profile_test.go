package service_test

import (
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/sharedlib/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type GetProfileServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mocks.UserRepositoryMock
	eventEmitter       *mocks.EmitterMock
	fileService        *mocks.MockFileServiceClient
	notificationStream *mocks.MockNatsInfra
	redisRepository    *mocks.MockRedisRepository
}

func (g *GetProfileServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	mockFileService := new(mocks.MockFileServiceClient)
	mockNotificationStream := new(mocks.MockNatsInfra)
	mockRedisRepository := new(mocks.MockRedisRepository)

	logger := zerolog.Nop()
	g.userRepository = mockUserRepo
	g.eventEmitter = mockEventEmitter
	g.fileService = mockFileService
	g.notificationStream = mockNotificationStream
	g.redisRepository = mockRedisRepository
	g.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter)
}

func (g *GetProfileServiceSuite) SetupTest() {
	g.userRepository.ExpectedCalls = nil
	g.eventEmitter.ExpectedCalls = nil
	g.fileService.ExpectedCalls = nil
	g.notificationStream.ExpectedCalls = nil
	g.redisRepository.ExpectedCalls = nil

	g.userRepository.Calls = nil
	g.eventEmitter.Calls = nil
	g.fileService.Calls = nil
	g.notificationStream.Calls = nil
	g.redisRepository.Calls = nil
}

func TestGetProfileServiceSuite(t *testing.T) {
	suite.Run(t, &GetProfileServiceSuite{})
}
func (g *GetProfileServiceSuite) TestUserService_GetProfile_Success() {
	userId := "user-123"
	image := "image.png"
	expectedUser := &model.User{
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Verified:         true,
		TwoFactorEnabled: true,
	}
	g.userRepository.On("QueryUserByUserId", userId).Return(expectedUser, nil)

	profile, err := g.userService.GetProfile(userId)

	g.NoError(err)
	g.Equal(expectedUser.FullName, profile.FullName)
	g.Equal(expectedUser.Image, profile.Image)
	g.Equal(expectedUser.Email, profile.Email)
	g.Equal(expectedUser.Verified, profile.Verified)
	g.Equal(expectedUser.TwoFactorEnabled, profile.TwoFactorEnabled)
	g.userRepository.AssertExpectations(g.T())
}

func (g *GetProfileServiceSuite) TestUserService_GetProfile_UserNotFound() {
	userId := "user-404"
	g.userRepository.On("QueryUserByUserId", userId).Return(nil, dto.Err_NOTFOUND_USER_NOT_FOUND)

	profile, err := g.userService.GetProfile(userId)

	g.Error(err)
	g.Empty(profile)
	g.userRepository.AssertExpectations(g.T())
}
