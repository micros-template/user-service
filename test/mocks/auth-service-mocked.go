package mocks

import (
	"context"

	"github.com/micros-template/proto-user/pkg/upb"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) CreateUser(user *upb.User) (*upb.Status, error) {
	args := m.Called(user)
	status, _ := args.Get(0).(*upb.Status)
	return status, args.Error(1)
}

func (m *MockAuthService) UpdateUser(ctx context.Context, user *upb.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthService) DeleteUser(ctx context.Context, user *upb.UserId) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
