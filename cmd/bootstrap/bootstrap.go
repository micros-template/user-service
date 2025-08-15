package bootstrap

import (
	"github.com/micros-template/user-service/cmd/di"
	"github.com/micros-template/user-service/config/env"

	"go.uber.org/dig"
)

func Run() *dig.Container {
	env.Load()
	container := di.BuildContainer()
	return container
}
