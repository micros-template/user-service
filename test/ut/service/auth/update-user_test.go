package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/service"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/proto-user/pkg/upb"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateUserAuthServiceSuite struct {
	suite.Suite
	authService    service.AuthService
	userRepository *mocks.UserRepositoryMock
	eventEmitter   *mocks.EmitterMock
}

func (u *UpdateUserAuthServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	logger := zerolog.Nop()
	u.userRepository = mockUserRepo
	u.eventEmitter = mockEventEmitter
	u.authService = service.NewAuthService(mockUserRepo, mockEventEmitter, logger)
}

func (u *UpdateUserAuthServiceSuite) SetupTest() {
	u.userRepository.ExpectedCalls = nil
	u.eventEmitter.ExpectedCalls = nil

	u.userRepository.Calls = nil
	u.eventEmitter.Calls = nil
}

func TestUpdateUserAuthServiceSuite(t *testing.T) {
	suite.Run(t, &UpdateUserAuthServiceSuite{})
}

func (u *UpdateUserAuthServiceSuite) TestAuthService_UpdateUser_Success() {
	image := "img.png"
	user := &upb.User{
		Id:               "user-123",
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Password:         "hashedpass",
		Verified:         true,
		TwoFactorEnabled: true,
	}
	u.userRepository.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(nil).Once()
	u.eventEmitter.On("UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User")).Return(nil).Once()

	err := u.authService.UpdateUser(context.TODO(), user)
	u.NoError(err)
	u.userRepository.AssertExpectations(u.T())
}

func (u *UpdateUserAuthServiceSuite) TestAuthService_UpdateUser_RepoError() {

	image := "img.png"
	user := &upb.User{
		Id:               "user-123",
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Password:         "hashedpass",
		Verified:         true,
		TwoFactorEnabled: true,
	}
	expectedErr := errors.New("db error")
	u.userRepository.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(expectedErr).Once()

	err := u.authService.UpdateUser(context.TODO(), user)
	u.ErrorIs(err, expectedErr)
	u.userRepository.AssertExpectations(u.T())
}

func (u *UpdateUserAuthServiceSuite) TestAuthService_UpdateUser_EventEmitterCalled() {
	image := "img.png"

	user := &upb.User{
		Id:               "user-123",
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Password:         "hashedpass",
		Verified:         true,
		TwoFactorEnabled: true,
	}
	u.userRepository.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(nil).Once()
	u.eventEmitter.On("UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User")).Return(nil).Once()

	err := u.authService.UpdateUser(context.TODO(), user)
	u.NoError(err)

	// Wait for goroutine to finish (in real code, use sync.WaitGroup or channel)
	time.Sleep(time.Second)
	u.eventEmitter.AssertCalled(u.T(), "UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User"))
}
