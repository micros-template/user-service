package service_test

import (
	"context"
	"testing"
	"time"

	"10.1.20.130/dropping/proto-user/pkg/upb"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DeleteUserAuthServiceSuite struct {
	suite.Suite
	authService    service.AuthService
	userRepository *mocks.UserRepositoryMock
	eventEmitter   *mocks.EmitterMock
}

func (d *DeleteUserAuthServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	logger := zerolog.Nop()
	d.userRepository = mockUserRepo
	d.eventEmitter = mockEventEmitter
	d.authService = service.NewAuthService(mockUserRepo, mockEventEmitter, logger)
}

func (d *DeleteUserAuthServiceSuite) SetupTest() {
	d.userRepository.ExpectedCalls = nil
	d.eventEmitter.ExpectedCalls = nil

	d.userRepository.Calls = nil
	d.eventEmitter.Calls = nil
}

func TestDeleteUserAuthServiceSuite(t *testing.T) {
	suite.Run(t, &DeleteUserAuthServiceSuite{})
}

func (d *DeleteUserAuthServiceSuite) TestAuthService_DeleteUser_Success() {
	u := &upb.UserId{
		UserId: "user-id-123",
	}
	d.userRepository.On("DeleteUser", mock.Anything).Return(nil)
	d.eventEmitter.On("DeleteUser", mock.Anything, u).Return(nil).Once()

	err := d.authService.DeleteUser(context.TODO(), u)

	d.NoError(err)

	d.userRepository.AssertExpectations(d.T())
	time.Sleep(time.Second)
	d.eventEmitter.AssertExpectations(d.T())
}
func (d *DeleteUserAuthServiceSuite) TestAuthService_DeleteUser_userNotFound() {
	u := &upb.UserId{
		UserId: "user-id-123",
	}
	d.userRepository.On("DeleteUser", mock.Anything).Return(dto.Err_NOTFOUND_USER_NOT_FOUND)

	err := d.authService.DeleteUser(context.TODO(), u)

	d.Error(err)
	d.userRepository.AssertExpectations(d.T())
}
