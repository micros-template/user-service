package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/internal/infrastructure/grpc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type HTTPServer struct {
	Container   *dig.Container
	ServerReady chan bool
	Address     string
}

func (s *HTTPServer) Run(ctx context.Context) {
	err := s.Container.Invoke(
		func(
			logger zerolog.Logger,
			router *gin.Engine,
			uh handler.UserHandler,
			pgx *pgxpool.Pool,
			nc *nats.Conn,
			redis *redis.Client,
			grpcClientManager *grpc.GRPCClientManager,

		) {
			defer grpcClientManager.CloseAllConnections()
			defer pgx.Close()
			defer func() {
				if err := redis.Close(); err != nil {
					logger.Error().Err(err).Msg("Failed to close Redis client")
				}
			}()
			defer func() {
				if err := nc.Drain(); err != nil {
					logger.Error().Err(err).Msg("Failed to drain nats client")
				}
			}()

			handler.RegisterUserRoutes(router, uh)
			srv := &http.Server{
				Addr:              s.Address,
				Handler:           router,
				ReadHeaderTimeout: 5 * time.Second,
			}
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal().Err(err).Msg("Failed to listen and serve http server")
				}
			}()

			if s.ServerReady != nil {
				for range 50 {
					conn, err := net.DialTimeout("tcp", s.Address, 100*time.Millisecond)
					if err == nil {
						if err := conn.Close(); err != nil {
							logger.Fatal().Err(err).Msg("establish check connection failed to close")
						}
						s.ServerReady <- true
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
			logger.Info().Msgf("HTTP Server Starting in port %s", s.Address)

			<-ctx.Done()
			logger.Info().Msg("Shutting down server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Fatal().Err(err).Msg("Server forced to shutdown")
			}
			logger.Info().Msg("Server exiting...")
		})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}
