package mocks

import (
	"github.com/micros-template/sharedlib/model"
	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) CreateNewUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) QueryUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *UserRepositoryMock) QueryUserByUserId(userId string) (*model.User, error) {
	args := m.Called(userId)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *UserRepositoryMock) UpdateUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) DeleteUser(userId string) error {
	args := m.Called(userId)
	return args.Error(0)
}
