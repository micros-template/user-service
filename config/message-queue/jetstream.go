package messagequeue

import (
	_mq "10.1.20.130/dropping/user-service/internal/infrastructure/message-queue"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NewJetstream(nc *nats.Conn) jetstream.JetStream {
	js, err := jetstream.New(nc)
	if err != nil {
		panic("failed to init jetstream")
	}
	_mq.NewNotificationStream(js)
	return js
}
