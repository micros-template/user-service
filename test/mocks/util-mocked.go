package mocks

import (
	"github.com/stretchr/testify/mock"
)

type LoggerServiceUtilMock struct {
	mock.Mock
}

func (l *LoggerServiceUtilMock) EmitLog(msgType, msg string) error {
	args := l.Called(msgType, msg)
	return args.Error(0)
}
