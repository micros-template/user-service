package mocks

import (
	"context"
	"time"

	m "github.com/stretchr/testify/mock"
)

type MockRedisRepository struct {
	m.Mock
}

func (m *MockRedisRepository) GetResource(ctx context.Context, s string) (string, error) {
	args := m.Called(ctx, s)
	return args.String(0), args.Error(1)
}

func (m *MockRedisRepository) SetResource(ctx context.Context, s1 string, s2 string, t time.Duration) error {
	args := m.Called(ctx, s1, s2, t)
	return args.Error(0)
}

func (m *MockRedisRepository) RemoveResource(ctx context.Context, s string) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}
