package service_test

import (
	"errors"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/service"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/proto-user/pkg/upb"
	"github.com/dropboks/sharedlib/model"
	"github.com/dropboks/sharedlib/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateUserServiceSuite struct {
	suite.Suite
	authService    service.AuthService
	userRepository *mocks.UserRepositoryMock
	eventEmitter   *mocks.EmitterMock
}

func (c *CreateUserServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	logger := zerolog.Nop()
	c.userRepository = mockUserRepo
	c.eventEmitter = mockEventEmitter
	c.authService = service.NewAuthService(mockUserRepo, mockEventEmitter, logger)
}

func (c *CreateUserServiceSuite) SetupTest() {
	c.userRepository.ExpectedCalls = nil
	c.eventEmitter.ExpectedCalls = nil

	c.userRepository.Calls = nil
	c.eventEmitter.Calls = nil
}

func TestCreateUserServiceSuite(t *testing.T) {
	suite.Run(t, &CreateUserServiceSuite{})
}

func (c *CreateUserServiceSuite) TestAuthService_CreateUser_Success() {
	image := "img.png"
	testUser := &upb.User{
		Id:               "123",
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Password:         "password",
		Verified:         true,
		TwoFactorEnabled: false,
	}
	expectedUser := &model.User{
		ID:               "123",
		FullName:         "John Doe",
		Image:            utils.StringPtr("img.png"),
		Email:            "john@example.com",
		Password:         "password",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	c.userRepository.On("CreateNewUser", expectedUser).Return(nil).Once()
	c.eventEmitter.On("InsertUser", mock.Anything, mock.MatchedBy(func(u *upb.User) bool {
		return u.GetId() == testUser.GetId() &&
			u.GetFullName() == testUser.GetFullName() &&
			u.GetImage() == testUser.GetImage() &&
			u.GetEmail() == testUser.GetEmail() &&
			u.GetPassword() == testUser.GetPassword() &&
			u.GetVerified() == testUser.GetVerified() &&
			u.GetTwoFactorEnabled() == testUser.GetTwoFactorEnabled()
	})).Return(nil).Maybe()

	status, err := c.authService.CreateUser(testUser)

	c.NoError(err)
	c.NotNil(status)
	c.True(status.Success)
	c.userRepository.AssertExpectations(c.T())
	// eventEmitter is called in a goroutine, so we wait a bit
	time.Sleep(time.Second)
	c.eventEmitter.AssertCalled(c.T(), "InsertUser", mock.Anything, mock.AnythingOfType("*upb.User"))
}

func (c *CreateUserServiceSuite) TestAuthService_CreateUser_RepositoryError() {
	image := "img.png"
	testUser := &upb.User{
		Id:               "123",
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
		Password:         "password",
		Verified:         true,
		TwoFactorEnabled: false,
	}
	expectedUser := &model.User{
		ID:               "123",
		FullName:         "John Doe",
		Image:            utils.StringPtr("img.png"),
		Email:            "john@example.com",
		Password:         "password",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	repoErr := errors.New("repo error")
	c.userRepository.On("CreateNewUser", expectedUser).Return(repoErr).Once()

	status, err := c.authService.CreateUser(testUser)

	c.Error(err)
	c.Nil(status)
	c.Equal(repoErr, err)
	c.userRepository.AssertExpectations(c.T())

	time.Sleep(time.Second)
	c.eventEmitter.AssertNotCalled(c.T(), "InsertUser", mock.Anything, mock.Anything)
}
