package service_test

import (
	"testing"
	"time"

	"github.com/micros-template/user-service/internal/domain/dto"
	"github.com/micros-template/user-service/internal/domain/service"
	mk "github.com/micros-template/user-service/test/mocks"

	"github.com/micros-template/sharedlib/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DeleteUserServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mk.UserRepositoryMock
	eventEmitter       *mk.EmitterMock
	fileService        *mk.MockFileServiceClient
	notificationStream *mk.MockNatsInfra
	redisRepository    *mk.MockRedisRepository
	logEmitter         *mk.LoggerInfraMock
}

func (d *DeleteUserServiceSuite) SetupSuite() {

	mockUserRepo := new(mk.UserRepositoryMock)
	mockEventEmitter := new(mk.EmitterMock)
	mockFileService := new(mk.MockFileServiceClient)
	mockNotificationStream := new(mk.MockNatsInfra)
	mockRedisRepository := new(mk.MockRedisRepository)
	mockLogEmitter := new(mk.LoggerInfraMock)

	logger := zerolog.Nop()
	d.userRepository = mockUserRepo
	d.eventEmitter = mockEventEmitter
	d.fileService = mockFileService
	d.notificationStream = mockNotificationStream
	d.redisRepository = mockRedisRepository
	d.logEmitter = mockLogEmitter
	d.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter, mockLogEmitter)
}

func (d *DeleteUserServiceSuite) SetupTest() {
	d.userRepository.ExpectedCalls = nil
	d.eventEmitter.ExpectedCalls = nil
	d.fileService.ExpectedCalls = nil
	d.notificationStream.ExpectedCalls = nil
	d.redisRepository.ExpectedCalls = nil
	d.logEmitter.ExpectedCalls = nil

	d.userRepository.Calls = nil
	d.eventEmitter.Calls = nil
	d.fileService.Calls = nil
	d.notificationStream.Calls = nil
	d.redisRepository.Calls = nil
	d.logEmitter.Calls = nil
}

func TestDeleteUserServiceSuite(t *testing.T) {
	suite.Run(t, &DeleteUserServiceSuite{})
}
func (d *DeleteUserServiceSuite) TestUserService_DeleteUser_Success() {
	u := model.User{
		ID:               "userid-123",
		FullName:         "test_user",
		Image:            new(string),
		Email:            "test@example.com",
		Password:         "$2a$10$Nwjs8PdFOCnjbRM3x/2WAuEtqOSrm6wHByYaw0ZDp5mV7e560dIb6",
		Verified:         true,
		TwoFactorEnabled: false,
	}
	req := dto.DeleteUserRequest{
		Password: "password123",
	}
	d.userRepository.On("QueryUserByUserId", "userid-123").Return(&u, nil)
	d.userRepository.On("DeleteUser", "userid-123").Return(nil)
	d.eventEmitter.On("DeleteUser", mock.Anything, mock.Anything).Return(nil).Once()
	err := d.userService.DeleteUser(&req, "userid-123")

	d.NoError(err)
	d.userRepository.AssertExpectations(d.T())

	time.Sleep(time.Second)
	d.eventEmitter.AssertExpectations(d.T())
}
func (d *DeleteUserServiceSuite) TestUserService_DeleteUser_UserNotFound() {
	req := dto.DeleteUserRequest{
		Password: "password123",
	}
	d.userRepository.On("QueryUserByUserId", "userid-123").Return(nil, dto.Err_NOTFOUND_USER_NOT_FOUND)

	err := d.userService.DeleteUser(&req, "userid-123")

	d.Error(err)
	d.userRepository.AssertExpectations(d.T())
}
func (d *DeleteUserServiceSuite) TestUserService_DeleteUser_WrongPassword() {
	u := model.User{
		ID:               "userid-123",
		FullName:         "test_user",
		Image:            new(string),
		Email:            "test@example.com",
		Password:         "$2a$10$Nwjs8PdFOCnjbRM3x/2WAuEtqOSrm6wHByYaw0ZDp5mV7e560dIb6",
		Verified:         true,
		TwoFactorEnabled: false,
	}
	req := dto.DeleteUserRequest{
		Password: "password1234",
	}
	d.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)
	d.userRepository.On("QueryUserByUserId", "userid-123").Return(&u, nil)

	err := d.userService.DeleteUser(&req, "userid-123")

	d.Error(err)
	d.userRepository.AssertExpectations(d.T())

	time.Sleep(time.Second)
	d.logEmitter.AssertExpectations(d.T())
}
