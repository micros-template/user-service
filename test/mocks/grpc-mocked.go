package mocks

import (
	"context"

	"10.1.20.130/dropping/proto-file/pkg/fpb"
	m "github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockFileServiceClient struct {
	m.Mock
}

func (m *MockFileServiceClient) SaveProfileImage(ctx context.Context, in *fpb.Image, opts ...grpc.CallOption) (*fpb.ImageName, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fpb.ImageName), args.Error(1)
}

func (m *MockFileServiceClient) RemoveProfileImage(ctx context.Context, in *fpb.ImageName, opts ...grpc.CallOption) (*fpb.Status, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fpb.Status), args.Error(1)
}
