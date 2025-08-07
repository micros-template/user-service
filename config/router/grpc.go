package router

import (
	"context"
	"encoding/json"
	"time"

	"10.1.20.130/dropping/log-management/pkg"
	ld "10.1.20.130/dropping/log-management/pkg/dto"
	"google.golang.org/grpc"
)

func loggingUnaryInterceptor(logEmitter pkg.LogEmitter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)
		if err != nil {
			logData := map[string]interface{}{
				"type":    "access",
				"status":  "error",
				"method":  info.FullMethod,
				"latency": elapsed.String(),
				"level":   "error",
			}
			logDataBytes, _ := json.Marshal(logData)
			logEmitter.EmitLog(ctx, ld.LogMessage{
				Type:     "ERR",
				Service:  "user_service",
				Msg:      string(logDataBytes),
				Protocol: "GRPC",
			})
		} else {
			logData := map[string]interface{}{
				"type":    "access",
				"status":  "success",
				"method":  info.FullMethod,
				"latency": elapsed.String(),
				"level":   "info",
			}
			logDataBytes, _ := json.Marshal(logData)
			logEmitter.EmitLog(ctx, ld.LogMessage{
				Type:     "INFO",
				Service:  "user_service",
				Msg:      string(logDataBytes),
				Protocol: "GRPC",
			})
		}
		return resp, err
	}
}

func NewGRPC(logEmitter pkg.LogEmitter) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor(logEmitter)),
	)
	return grpcServer
}
