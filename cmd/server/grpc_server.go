package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/dropboks/user-service/internal/domain/handler"
	"github.com/dropboks/user-service/internal/domain/service"
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
	) {
		defer db.Close()
		listen, err := net.Listen("tcp", s.Address)
		if err != nil {
			logger.Fatal().Msgf("failed to listen:%v", err)
		}
		handler.RegisterAuthService(grpcServer, svc)

		go func() {
			if serveErr := grpcServer.Serve(listen); serveErr != nil {
				logger.Fatal().Msgf("gRPC server error: %v", serveErr)
			}
		}()

		if s.ServerReady != nil {
			for range 50 {
				conn, err := net.DialTimeout("tcp", s.Address, 100*time.Millisecond)
				if err == nil {
					conn.Close()
					s.ServerReady <- true
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		logger.Info().Msg("gRPC server running in port " + s.Address)

		<-ctx.Done()
		logger.Info().Msg("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		logger.Info().Msg("gRPC server stopped gracefully.")
	})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}
