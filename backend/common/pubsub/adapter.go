package pubsub

import (
	"time"

	"cloud.google.com/go/pubsub"
)

type PubSubMessage interface {
	Ack()
	Nack()
	GetData() []byte
	GetPublishTime() time.Time
}

type PubSubMessageAdapter struct {
	Msg *pubsub.Message
}

func (a *PubSubMessageAdapter) Ack()                      { a.Msg.Ack() }
func (a *PubSubMessageAdapter) Nack()                     { a.Msg.Nack() }
func (a *PubSubMessageAdapter) GetData() []byte           { return a.Msg.Data }
func (a *PubSubMessageAdapter) GetPublishTime() time.Time { return a.Msg.PublishTime }
