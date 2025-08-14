package handler_test

import (
	"context"
	"testing"

	"10.1.20.130/dropping/proto-user/pkg/upb"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type DeleteUserHandlerSuite struct {
	suite.Suite
	authHandler     handler.AuthGrpcHandler
	mockAuthService *mocks.MockAuthService
}

func (d *DeleteUserHandlerSuite) SetupSuite() {
	mockedAuthService := new(mocks.MockAuthService)
	d.mockAuthService = mockedAuthService

	grpcServer := grpc.NewServer()
	handler.RegisterAuthService(grpcServer, mockedAuthService)
	d.authHandler = *handler.NewAuthGrpcHandler(mockedAuthService)
}

func (d *DeleteUserHandlerSuite) SetupTest() {
	d.mockAuthService.ExpectedCalls = nil
	d.mockAuthService.Calls = nil
}

func TestDeleteUserHandlerSuite(t *testing.T) {
	suite.Run(t, &DeleteUserHandlerSuite{})
}
func (d *DeleteUserHandlerSuite) TestAuthHandler_DeleteUserHandler_Success() {
	u := &upb.UserId{
		UserId: "user-id-123",
	}
	expectedStatus := &upb.Status{
		Success: true,
	}
	d.mockAuthService.On("DeleteUser", mock.Anything, u).Return(nil)
	s, err := d.authHandler.DeleteUser(context.Background(), u)

	d.NotNil(s)
	d.Equal(expectedStatus, s)
	d.NoError(err)
}

func (d *DeleteUserHandlerSuite) TestAuthHandler_DeleteUserHandler_UserNotFound() {
	u := &upb.UserId{
		UserId: "user-id-123",
	}
	d.mockAuthService.On("DeleteUser", mock.Anything, u).Return(dto.Err_NOTFOUND_USER_NOT_FOUND)
	s, err := d.authHandler.DeleteUser(context.Background(), u)

	d.Nil(s)
	d.Error(err)
}

func (d *DeleteUserHandlerSuite) TestAuthHandler_DeleteUserHandler_InternalServerError() {
	u := &upb.UserId{
		UserId: "user-id-123",
	}
	d.mockAuthService.On("DeleteUser", mock.Anything, u).Return(dto.Err_INTERNAL_FAILED_BUILD_QUERY)
	s, err := d.authHandler.DeleteUser(context.Background(), u)

	d.Nil(s)
	d.Error(err)
}
