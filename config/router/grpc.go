package router

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func loggingUnaryInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		logger.Info().
			Str("method", info.FullMethod).
			Msg("gRPC request received")
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)
		if err != nil {
			logger.Error().
				Str("method", info.FullMethod).
				Int64("duration_ms", elapsed.Milliseconds()).
				Err(err).
				Msg("gRPC request error")
		} else {
			logger.Info().
				Str("method", info.FullMethod).
				Int64("duration_ms", elapsed.Milliseconds()).
				Msg("gRPC request completed")
		}
		return resp, err
	}
}

func NewGRPC(logger zerolog.Logger) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor(logger)),
	)
	return grpcServer
}
