package main

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
	msg *pubsub.Message
}

func (a *PubSubMessageAdapter) Ack()                      { a.msg.Ack() }
func (a *PubSubMessageAdapter) Nack()                     { a.msg.Nack() }
func (a *PubSubMessageAdapter) GetData() []byte           { return a.msg.Data }
func (a *PubSubMessageAdapter) GetPublishTime() time.Time { return a.msg.PublishTime }
