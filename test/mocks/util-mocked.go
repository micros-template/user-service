package mocks

import (
	"10.1.20.130/dropping/log-management/pkg"
	"github.com/stretchr/testify/mock"
)

type UserServiceUtilMock struct {
	mock.Mock
}

func (m *UserServiceUtilMock) EmitLog(logEmitter pkg.LogEmitter, msgType, msg string) error {
	args := m.Called(logEmitter, msgType, msg)
	return args.Error(0)
}
