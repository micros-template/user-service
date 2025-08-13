package service_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"10.1.20.130/dropping/log-management/pkg/mocks"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	mk "10.1.20.130/dropping/user-service/test/mocks"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateEmailServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mk.UserRepositoryMock
	eventEmitter       *mk.EmitterMock
	fileService        *mk.MockFileServiceClient
	notificationStream *mk.MockNatsInfra
	redisRepository    *mk.MockRedisRepository
	mockUtil           *mk.LoggerServiceUtilMock
}

func (u *UpdateEmailServiceSuite) SetupSuite() {

	mockUserRepo := new(mk.UserRepositoryMock)
	mockEventEmitter := new(mk.EmitterMock)
	mockFileService := new(mk.MockFileServiceClient)
	mockNotificationStream := new(mk.MockNatsInfra)
	mockRedisRepository := new(mk.MockRedisRepository)
	mockUserServiceUtil := new(mk.LoggerServiceUtilMock)
	mockLogEmitter := new(mocks.LogEmitterMock)

	logger := zerolog.Nop()
	u.userRepository = mockUserRepo
	u.eventEmitter = mockEventEmitter
	u.fileService = mockFileService
	u.notificationStream = mockNotificationStream
	u.redisRepository = mockRedisRepository
	u.mockUtil = mockUserServiceUtil
	u.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter, mockLogEmitter, mockUserServiceUtil)
}

func (u *UpdateEmailServiceSuite) SetupTest() {
	u.userRepository.ExpectedCalls = nil
	u.eventEmitter.ExpectedCalls = nil
	u.fileService.ExpectedCalls = nil
	u.notificationStream.ExpectedCalls = nil
	u.redisRepository.ExpectedCalls = nil
	u.mockUtil.ExpectedCalls = nil

	u.userRepository.Calls = nil
	u.eventEmitter.Calls = nil
	u.fileService.Calls = nil
	u.notificationStream.Calls = nil
	u.redisRepository.Calls = nil
	u.mockUtil.Calls = nil
}

func TestUpdateEmailServiceSuite(t *testing.T) {
	suite.Run(t, &UpdateEmailServiceSuite{})
}

func (u *UpdateEmailServiceSuite) TestUserService_UpdateEmail_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	req := &dto.UpdateEmailRequest{
		Email: email,
	}
	userId := "user-123"

	u.redisRepository.On("SetResource", mock.Anything, "newEmail:"+userId, req.Email, mock.Anything).Return(nil).Once()
	u.redisRepository.On("SetResource", mock.Anything, "changeEmailToken:"+userId, mock.Anything, mock.Anything).Return(nil).Once()
	u.notificationStream.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(&jetstream.PubAck{}, nil).Once()

	err := u.userService.UpdateEmail(req, userId)

	u.NoError(err)
	u.redisRepository.AssertExpectations(u.T())
	u.notificationStream.AssertExpectations(u.T())
}

func (u *UpdateEmailServiceSuite) TestUserService_UpdateEmail_RedisError() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	req := &dto.UpdateEmailRequest{
		Email: email,
	}
	userId := "user-123"

	u.redisRepository.On("SetResource", mock.Anything, "newEmail:"+userId, req.Email, mock.Anything).Return(dto.Err_INTERNAL_SET_RESOURCE).Once()

	err := u.userService.UpdateEmail(req, userId)

	u.Error(err)
	u.redisRepository.AssertExpectations(u.T())
}

func (u *UpdateEmailServiceSuite) TestUserService_UpdateEmail_JetstreamError() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	req := &dto.UpdateEmailRequest{
		Email: email,
	}
	userId := "user-123"

	u.redisRepository.On("SetResource", mock.Anything, "newEmail:"+userId, req.Email, mock.Anything).Return(nil).Once()
	u.redisRepository.On("SetResource", mock.Anything, "changeEmailToken:"+userId, mock.Anything, mock.Anything).Return(nil).Once()
	u.notificationStream.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(&jetstream.PubAck{}, errors.New("failed to update email")).Once()
	u.mockUtil.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userService.UpdateEmail(req, userId)

	u.Error(err)
	u.redisRepository.AssertExpectations(u.T())
	u.notificationStream.AssertExpectations(u.T())

	time.Sleep(time.Second)
	u.mockUtil.AssertExpectations(u.T())
}
