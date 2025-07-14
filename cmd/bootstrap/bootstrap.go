package bootstrap

import (
	"github.com/dropboks/user-service/cmd/di"
	"github.com/dropboks/user-service/config/env"
	"go.uber.org/dig"
)

func Run() *dig.Container {
	env.Load()
	container := di.BuildContainer()
	return container
}
