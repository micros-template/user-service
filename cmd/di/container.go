package di

import (
	"github.com/micros-template/user-service/config/cache"
	db "github.com/micros-template/user-service/config/database"
	logemitter "github.com/micros-template/user-service/config/log_emitter"
	"github.com/micros-template/user-service/config/logger"
	mq "github.com/micros-template/user-service/config/message-queue"
	"github.com/micros-template/user-service/config/router"
	"github.com/micros-template/user-service/internal/domain/handler"
	"github.com/micros-template/user-service/internal/domain/repository"
	"github.com/micros-template/user-service/internal/domain/service"
	_cache "github.com/micros-template/user-service/internal/infrastructure/cache"
	_db "github.com/micros-template/user-service/internal/infrastructure/database"
	"github.com/micros-template/user-service/internal/infrastructure/eventbus"
	"github.com/micros-template/user-service/internal/infrastructure/grpc"
	_logger "github.com/micros-template/user-service/internal/infrastructure/logger"
	_mq "github.com/micros-template/user-service/internal/infrastructure/message-queue"

	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()
	// logger
	if err := container.Provide(logger.New); err != nil {
		panic("Failed to provide logger: " + err.Error())
	}
	// db connection
	if err := container.Provide(db.New); err != nil {
		panic("Failed to provide database: " + err.Error())
	}
	// db querier
	if err := container.Provide(_db.NewQuerier); err != nil {
		panic("Failed to provide database querier`: " + err.Error())
	}
	// nats connection
	if err := container.Provide(mq.New); err != nil {
		panic("Failed to provide nats connection: " + err.Error())
	}
	// nats connection
	if err := container.Provide(mq.NewJetstream); err != nil {
		panic("Failed to provide jetstream instance: " + err.Error())
	}
	// nats infrastructure
	if err := container.Provide(_mq.NewNatsInfrastructure); err != nil {
		panic("Failed to provide nats infrastructure: " + err.Error())
	}
	// log emitter
	if err := container.Provide(logemitter.NewInfraLogEmitter); err != nil {
		panic("Failed to provide log emitter: " + err.Error())
	}
	// event emitter
	if err := container.Provide(eventbus.NewEventBusEmitterInfra); err != nil {
		panic("Failed to provide event bus emitter: " + err.Error())
	}
	// redis connection
	if err := container.Provide(cache.New); err != nil {
		panic("Failed to provide cache client: " + err.Error())
	}
	// redis infra
	if err := container.Provide(_cache.New); err != nil {
		panic("Failed to provide cache infrastructure: " + err.Error())
	}
	// grpc client manager
	if err := container.Provide(grpc.NewGRPCClientManager); err != nil {
		panic("Failed to provide GRPC Client Manager: " + err.Error())
	}
	// file_service connection
	if err := container.Provide(grpc.NewFileServiceConnection); err != nil {
		panic("Failed to provide user service grpc connection: " + err.Error())
	}
	// user service utils
	if err := container.Provide(_logger.NewLoggerInfra); err != nil {
		panic("Failed to provide user service utils: " + err.Error())
	}
	// user_repo
	if err := container.Provide(repository.NewUserRepository); err != nil {
		panic("Failed to provide authRepository: " + err.Error())
	}
	// redis_repo
	if err := container.Provide(repository.NewRedisRepository); err != nil {
		panic("Failed to provide cache client: " + err.Error())
	}
	// auth_service
	if err := container.Provide(service.NewAuthService); err != nil {
		panic("Failed to provide auth service: " + err.Error())
	}
	// user_service
	if err := container.Provide(service.NewUserService); err != nil {
		panic("Failed to provide user service: " + err.Error())
	}
	// user_handler
	if err := container.Provide(handler.NewUserHandler); err != nil {
		panic("Failed to provide user handler: " + err.Error())
	}
	if err := container.Provide(router.NewHTTP); err != nil {
		panic("Failed to provide HTTP Server: " + err.Error())
	}
	if err := container.Provide(router.NewGRPC); err != nil {
		panic("Failed to provide gRPC Server: " + err.Error())
	}
	return container
}
