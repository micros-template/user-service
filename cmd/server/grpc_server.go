package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/micros-template/user-service/internal/domain/handler"
	"github.com/micros-template/user-service/internal/domain/service"
	"github.com/micros-template/user-service/internal/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	Container   *dig.Container
	ServerReady chan bool
	Address     string
}

func (s *GRPCServer) Run(ctx context.Context) {
	err := s.Container.Invoke(func(
		grpcServer *grpc.Server,
		logger zerolog.Logger,
		db *pgxpool.Pool,
		svc service.AuthService,
		logEmitter logger.LoggerInfra,

	) {
		defer db.Close()
		listen, err := net.Listen("tcp", s.Address)
		if err != nil {
			logger.Fatal().Msgf("failed to listen:%v", err)
		}
		handler.RegisterAuthService(grpcServer, svc)

		go func() {
			if serveErr := grpcServer.Serve(listen); serveErr != nil {
				go func() {
					if err := logEmitter.EmitLog("ERR", fmt.Sprintf("failed to listen an serve gRPC server:%v", err)); err != nil {
						logger.Error().Err(err).Msg("failed to emit log")
					}
				}()
				logger.Fatal().Msgf("failed to listen an serve gRPC server: %v", serveErr)
			}
		}()

		if s.ServerReady != nil {
			for range 50 {
				conn, err := net.DialTimeout("tcp", s.Address, 100*time.Millisecond)
				if err == nil {
					if err := conn.Close(); err != nil {
						go func() {
							if err := logEmitter.EmitLog("ERR", fmt.Sprintf("establish check connection failed to close:%v", err)); err != nil {
								logger.Error().Err(err).Msg("failed to emit log")
							}
						}()
					}
					s.ServerReady <- true
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}

		go func() {
			if err := logEmitter.EmitLog("INFO", fmt.Sprintf("gRPC server running in port %s", s.Address)); err != nil {
				logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		logger.Info().Msg("gRPC server running in port " + s.Address)
		<-ctx.Done()

		go func() {
			if err := logEmitter.EmitLog("INFO", "Shutting down gRPC server..."); err != nil {
				logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		logger.Info().Msg("Shutting down gRPC server...")
		grpcServer.GracefulStop()

		go func() {
			if err := logEmitter.EmitLog("INFO", "gRPC server stopped gracefully."); err != nil {
				logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		logger.Info().Msg("gRPC server stopped gracefully.")
	})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}
