package mocks

import (
	"github.com/stretchr/testify/mock"
)

type LoggerInfraMock struct {
	mock.Mock
}

func (l *LoggerInfraMock) EmitLog(msgType, msg string) error {
	args := l.Called(msgType, msg)
	return args.Error(0)
}
