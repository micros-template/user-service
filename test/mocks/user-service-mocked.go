package mocks

import (
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"github.com/stretchr/testify/mock"
)

type UserServiceMock struct {
	mock.Mock
}

func (m *UserServiceMock) GetProfile(userId string) (dto.GetProfileResponse, error) {
	args := m.Called(userId)
	return args.Get(0).(dto.GetProfileResponse), args.Error(1)
}

func (m *UserServiceMock) UpdateUser(req *dto.UpdateUserRequest, userId string) error {
	args := m.Called(req, userId)
	return args.Error(0)
}

func (m *UserServiceMock) UpdateEmail(req *dto.UpdateEmailRequest, userId string) error {
	args := m.Called(req, userId)
	return args.Error(0)
}

func (m *UserServiceMock) UpdatePassword(req *dto.UpdatePasswordRequest, userId string) error {
	args := m.Called(req, userId)
	return args.Error(0)
}

func (m *UserServiceMock) DeleteUser(req *dto.DeleteUserRequest, userId string) error {
	args := m.Called(req, userId)
	return args.Error(0)
}
