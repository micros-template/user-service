package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dropboks/user-service/cmd/bootstrap"
	"github.com/dropboks/user-service/cmd/server"
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
