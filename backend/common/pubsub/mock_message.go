package pubsub

import "time"

type FakeMessage struct {
	Data        []byte
	PublishTime time.Time
	Acked       bool
	Nacked      bool
}

func (m *FakeMessage) Ack()                      { m.Acked = true }
func (m *FakeMessage) Nack()                     { m.Nacked = true }
func (m *FakeMessage) GetData() []byte           { return m.Data }
func (m *FakeMessage) GetPublishTime() time.Time { return m.PublishTime }
