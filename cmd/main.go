// @title User Service API
// @version 1.0
// @description User service API for operation related to user
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1/user

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"10.1.20.130/dropping/user-service/cmd/bootstrap"
	"10.1.20.130/dropping/user-service/cmd/server"
	"github.com/spf13/viper"
)

func main() {
	container := bootstrap.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServerReady := make(chan bool)
	httpServerDone := make(chan struct{})
	httpServer := &server.HTTPServer{
		Container:   container,
		ServerReady: httpServerReady,
		Address:     ":" + viper.GetString("app.http.port"),
	}
	go func() {
		httpServer.Run(ctx)
		close(httpServerDone)
	}()
	<-httpServerReady

	grpcServerReady := make(chan bool)
	grpcServerDone := make(chan struct{})
	grpcServer := &server.GRPCServer{
		Container:   container,
		ServerReady: grpcServerReady,
		Address:     ":" + viper.GetString("app.grpc.port"),
	}
	go func() {
		grpcServer.Run(ctx)
		close(grpcServerDone)
	}()
	<-grpcServerReady

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)

	<-sig
	cancel()

	<-httpServerDone
	<-grpcServerDone
}
