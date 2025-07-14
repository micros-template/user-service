package handler_test

import (
	"context"
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/proto-user/pkg/upb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

type UpdateUserHandlerSuite struct {
	suite.Suite
	authHandler     handler.AuthGrpcHandler
	mockAuthService *mocks.MockAuthService
}

func (c *UpdateUserHandlerSuite) SetupSuite() {
	mockedAuthService := new(mocks.MockAuthService)
	c.mockAuthService = mockedAuthService

	grpcServer := grpc.NewServer()
	handler.RegisterAuthService(grpcServer, mockedAuthService)
	c.authHandler = *handler.NewAuthGrpcHandler(mockedAuthService)
}

func (c *UpdateUserHandlerSuite) SetupTest() {
	c.mockAuthService.ExpectedCalls = nil
	c.mockAuthService.Calls = nil
}

func TestUpdateUserHandlerSuite(t *testing.T) {
	suite.Run(t, &UpdateUserHandlerSuite{})
}
func (c *UpdateUserHandlerSuite) TestAuthHandler_UpdateUserHandler_Success() {
	ctx := context.Background()
	user := &upb.User{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedStatus := &upb.Status{
		Success: true,
	}

	c.mockAuthService.On("UpdateUser", mock.Anything, user).Return(nil)

	status, err := c.authHandler.UpdateUser(ctx, user)

	c.NoError(err)
	c.Equal(expectedStatus, status)
	c.mockAuthService.AssertExpectations(c.T())
}

func (c *UpdateUserHandlerSuite) TestAuthHandler_UpdateUserHandler_UserNotFound() {
	ctx := context.Background()
	user := &upb.User{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedError := dto.Err_NOTFOUND_USER_NOT_FOUND

	c.mockAuthService.On("UpdateUser", mock.Anything, user).Return(expectedError)

	status, err := c.authHandler.UpdateUser(ctx, user)

	c.Nil(status)
	c.Error(err)
	expectedGrpcErr := grpcStatus.Error(codes.NotFound, expectedError.Error())
	c.Equal(expectedGrpcErr.Error(), err.Error())
	c.mockAuthService.AssertExpectations(c.T())
}

func (c *UpdateUserHandlerSuite) TestAuthHandler_UpdateUserHandler_Error() {
	ctx := context.Background()
	user := &upb.User{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedError := dto.Err_INTERNAL_FAILED_INSERT_USER

	c.mockAuthService.On("UpdateUser", mock.Anything, user).Return(expectedError)

	status, err := c.authHandler.UpdateUser(ctx, user)

	c.Nil(status)
	c.Error(err)
	expectedGrpcErr := grpcStatus.Error(codes.Internal, expectedError.Error())
	c.Equal(expectedGrpcErr.Error(), err.Error())
	c.mockAuthService.AssertExpectations(c.T())
}
