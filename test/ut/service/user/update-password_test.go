package service_test

import (
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/sharedlib/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdatePasswordServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mocks.UserRepositoryMock
	eventEmitter       *mocks.EmitterMock
	fileService        *mocks.MockFileServiceClient
	notificationStream *mocks.MockNatsInfra
	redisRepository    *mocks.MockRedisRepository
}

func (u *UpdatePasswordServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	mockFileService := new(mocks.MockFileServiceClient)
	mockNotificationStream := new(mocks.MockNatsInfra)
	mockRedisRepository := new(mocks.MockRedisRepository)

	logger := zerolog.Nop()
	u.userRepository = mockUserRepo
	u.eventEmitter = mockEventEmitter
	u.fileService = mockFileService
	u.notificationStream = mockNotificationStream
	u.redisRepository = mockRedisRepository
	u.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter)
}

func (u *UpdatePasswordServiceSuite) SetupTest() {
	u.userRepository.ExpectedCalls = nil
	u.eventEmitter.ExpectedCalls = nil
	u.fileService.ExpectedCalls = nil
	u.notificationStream.ExpectedCalls = nil
	u.redisRepository.ExpectedCalls = nil

	u.userRepository.Calls = nil
	u.eventEmitter.Calls = nil
	u.fileService.Calls = nil
	u.notificationStream.Calls = nil
	u.redisRepository.Calls = nil
}

func TestUpdatePasswordServiceSuite(t *testing.T) {
	suite.Run(t, &UpdatePasswordServiceSuite{})
}

func (u *UpdatePasswordServiceSuite) TestUserService_UpdatePassword_Success() {
	userId := "user-123"
	oldPassword := "$2a$10$Nwjs8PdFOCnjbRM3x/2WAuEtqOSrm6wHByYaw0ZDp5mV7e560dIb6"
	newPassword := "new-password"

	req := &dto.UpdatePasswordRequest{
		Password:           "password123",
		NewPassword:        newPassword,
		ConfirmNewPassword: newPassword,
	}

	user := &model.User{
		ID:       userId,
		Password: oldPassword,
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(user, nil)
	u.userRepository.On("UpdateUser", mock.Anything).Return(nil)
	u.eventEmitter.On("UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User")).Return(nil).Maybe()

	err := u.userService.UpdatePassword(req, userId)

	u.NoError(err)
	u.userRepository.AssertExpectations(u.T())
	time.Sleep(time.Second)
	u.eventEmitter.AssertCalled(u.T(), "UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User"))
}

func (u *UpdatePasswordServiceSuite) TestUserService_UpdatePassword_UserNotFound() {
	userId := "user-123"
	newPassword := "new-password"

	req := &dto.UpdatePasswordRequest{
		Password:           "password123",
		NewPassword:        newPassword,
		ConfirmNewPassword: newPassword,
	}

	u.userRepository.On("QueryUserByUserId", userId).Return(nil, dto.Err_NOTFOUND_USER_NOT_FOUND).Once()
	err := u.userService.UpdatePassword(req, userId)

	u.Error(err)
	u.userRepository.AssertExpectations(u.T())
}
