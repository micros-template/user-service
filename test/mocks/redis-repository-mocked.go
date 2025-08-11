package mocks

import (
	"context"
	"time"

	m "github.com/stretchr/testify/mock"
)

type MockRedisRepository struct {
	m.Mock
}

func (m *MockRedisRepository) SetResource(ctx context.Context, s1 string, s2 string, t time.Duration) error {
	args := m.Called(ctx, s1, s2, t)
	return args.Error(0)
}
