package logger

import (
	"context"

	"10.1.20.130/dropping/log-management/pkg"
	ld "10.1.20.130/dropping/log-management/pkg/dto"
)

type LoggerInfra interface {
	EmitLog(msgType, msg string) error
}

type loggerInfra struct {
	logEmitter pkg.LogEmitter
}

func NewLoggerInfra(logEmitter pkg.LogEmitter) LoggerInfra {
	return &loggerInfra{
		logEmitter: logEmitter,
	}
}

func (u *loggerInfra) EmitLog(msgType, msg string) error {
	if err := u.logEmitter.EmitLog(context.Background(), ld.LogMessage{
		Type:     msgType,
		Service:  "user_service",
		Msg:      msg,
		Protocol: "SYSTEM",
	}); err != nil {
		return err
	}
	return nil
}
