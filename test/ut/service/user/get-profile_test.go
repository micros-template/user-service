package service_test

import (
	"testing"

	"10.1.20.130/dropping/log-management/pkg/mocks"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	mk "10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/sharedlib/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type GetProfileServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mk.UserRepositoryMock
	eventEmitter       *mk.EmitterMock
	fileService        *mk.MockFileServiceClient
	notificationStream *mk.MockNatsInfra
	redisRepository    *mk.MockRedisRepository
	mockUtil           *mk.UserServiceUtilMock
}

func (g *GetProfileServiceSuite) SetupSuite() {

	mockUserRepo := new(mk.UserRepositoryMock)
	mockEventEmitter := new(mk.EmitterMock)
	mockFileService := new(mk.MockFileServiceClient)
	mockNotificationStream := new(mk.MockNatsInfra)
	mockRedisRepository := new(mk.MockRedisRepository)
	mockUserServiceUtil := new(mk.UserServiceUtilMock)
	mockLogEmitter := new(mocks.LogEmitterMock)

	logger := zerolog.Nop()
	g.userRepository = mockUserRepo
	g.eventEmitter = mockEventEmitter
	g.fileService = mockFileService
	g.notificationStream = mockNotificationStream
	g.redisRepository = mockRedisRepository
	g.mockUtil = mockUserServiceUtil
	g.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter, mockLogEmitter, mockUserServiceUtil)
}

func (g *GetProfileServiceSuite) SetupTest() {
	g.userRepository.ExpectedCalls = nil
	g.eventEmitter.ExpectedCalls = nil
	g.fileService.ExpectedCalls = nil
	g.notificationStream.ExpectedCalls = nil
	g.redisRepository.ExpectedCalls = nil
	g.mockUtil.ExpectedCalls = nil

	g.userRepository.Calls = nil
	g.eventEmitter.Calls = nil
	g.fileService.Calls = nil
	g.notificationStream.Calls = nil
	g.redisRepository.Calls = nil
	g.mockUtil.Calls = nil
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
