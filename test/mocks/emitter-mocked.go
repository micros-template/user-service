package mocks

import (
	"context"

	"10.1.20.130/dropping/proto-user/pkg/upb"
	"github.com/stretchr/testify/mock"
)

type EmitterMock struct {
	mock.Mock
}

func (m *EmitterMock) InsertUser(ctx context.Context, user *upb.User) {
	m.Called(ctx, user)
}

func (m *EmitterMock) UpdateUser(ctx context.Context, user *upb.User) {
	m.Called(ctx, user)
}

func (m *EmitterMock) DeleteUser(ctx context.Context, user *upb.UserId) {
	m.Called(ctx, user)
}
