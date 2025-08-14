package handler

import (
	"context"

	upb "10.1.20.130/dropping/proto-user/pkg/upb"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_status "google.golang.org/grpc/status"
)

type AuthGrpcHandler struct {
	authService service.AuthService
	upb.UnimplementedUserServiceServer
}

func NewAuthGrpcHandler(authService service.AuthService) *AuthGrpcHandler {
	return &AuthGrpcHandler{
		authService: authService,
	}
}

func RegisterAuthService(grpc *grpc.Server, authService service.AuthService) {
	grpcHandler := NewAuthGrpcHandler(authService)
	upb.RegisterUserServiceServer(grpc, grpcHandler)
}

func (a *AuthGrpcHandler) CreateUser(c context.Context, user *upb.User) (*upb.Status, error) {
	status, err := a.authService.CreateUser(user)
	if err != nil {
		if err == dto.Err_INTERNAL_FAILED_BUILD_QUERY || err == dto.Err_INTERNAL_FAILED_INSERT_USER {
			return nil, _status.Error(codes.Internal, err.Error())
		}
		return nil, _status.Error(codes.Internal, err.Error())
	}
	return status, nil
}

func (a *AuthGrpcHandler) UpdateUser(c context.Context, user *upb.User) (*upb.Status, error) {
	if err := a.authService.UpdateUser(c, user); err != nil {
		switch err {
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			return nil, _status.Error(codes.NotFound, err.Error())
		case dto.Err_INTERNAL_FAILED_BUILD_QUERY, dto.Err_INTERNAL_FAILED_INSERT_USER:
			return nil, _status.Error(codes.Internal, err.Error())
		}
		return nil, _status.Error(codes.Internal, err.Error())
	}
	return &upb.Status{Success: true}, nil
}

func (a *AuthGrpcHandler) DeleteUser(c context.Context, user *upb.UserId) (*upb.Status, error) {
	if err := a.authService.DeleteUser(c, user); err != nil {
		switch err {
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			return nil, _status.Error(codes.NotFound, err.Error())
		case dto.Err_INTERNAL_FAILED_BUILD_QUERY, dto.Err_INTERNAL_FAILED_DELETE_USER:
			return nil, _status.Error(codes.Internal, err.Error())
		}
		return nil, _status.Error(codes.Internal, err.Error())
	}
	return &upb.Status{Success: true}, nil
}
