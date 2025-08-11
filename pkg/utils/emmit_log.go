package utils

import (
	"context"

	"10.1.20.130/dropping/log-management/pkg"
	ld "10.1.20.130/dropping/log-management/pkg/dto"
)

type UserServiceUtil interface {
	EmitLog(logEmitter pkg.LogEmitter, msgType, msg string) error
}

type userServiceUtil struct{}

func NewUserServiceUtil() UserServiceUtil {
	return &userServiceUtil{}
}

func (u *userServiceUtil) EmitLog(logEmitter pkg.LogEmitter, msgType, msg string) error {
	if err := logEmitter.EmitLog(context.Background(), ld.LogMessage{
		Type:     msgType,
		Service:  "user_service",
		Msg:      msg,
		Protocol: "SYSTEM",
	}); err != nil {
		return err
	}
	return nil
}
