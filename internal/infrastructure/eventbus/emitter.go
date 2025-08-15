package eventbus

import (
	"github.com/micros-template/event-bus-client/pkg/event"

	"github.com/micros-template/log-service/pkg"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewEventBusEmitterInfra(js jetstream.JetStream, logEmitter pkg.LogEmitter, logger zerolog.Logger) event.Emitter {
	sen := "user_service"
	sn := viper.GetString("jetstream.event.stream.name")
	sd := viper.GetString("jetstream.event.stream.description")
	gs := viper.GetString("jetstream.event.subject.global")
	sp := viper.GetString("jetstream.event.subject.event_bus")
	em := event.NewEmitter(js, logEmitter, sen, sn, sd, gs, sp, logger)
	return em
}
