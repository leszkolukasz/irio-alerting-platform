package main

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"alerting-platform/common/pubsub"

	email "notifier/email"
)

func TestHandleMessage_Notify_Success(t *testing.T) {
	mailer := &email.MockMailer{}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"oncaller": "admin@example.com", "incident_id": "INC-123", "service_id": 99}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.NotifyOncallerTopic, mailer)

	assert.True(t, mailer.SendCalled, "Mailer should be called")
	assert.Equal(t, "admin@example.com", mailer.LastTo)
	assert.Equal(t, "INC-123", mailer.LastIncidentID)
	assert.Equal(t, uint64(99), mailer.LastServiceID)

	assert.True(t, msg.Acked, "Message should be ACKed")
	assert.False(t, msg.Nacked, "Message should NOT be NACKed")
}

func TestHandleMessage_EmailError_StillAcks(t *testing.T) {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalOutput)

	mailer := &email.MockMailer{
		Err: errors.New("smtp timeout"),
	}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"oncaller": "admin@example.com", "incident_id": "INC-1", "service_id": 1}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.NotifyOncallerTopic, mailer)

	assert.True(t, mailer.SendCalled, "Mailer should try to send email even if it fails")

	assert.True(t, msg.Acked, "Message should be ACKed even on email error")
	assert.False(t, msg.Nacked)
}

func TestHandleMessage_InvalidJSON_DropsMessage(t *testing.T) {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalOutput)

	mailer := &email.MockMailer{}
	msg := &pubsub.FakeMessage{
		Data:        []byte(`{broken-json`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.NotifyOncallerTopic, mailer)

	assert.False(t, mailer.SendCalled, "Mailer should NOT be called for invalid JSON")
	assert.True(t, msg.Acked, "Invalid message should be ACKed (dropped)")
}

func TestHandleMessage_UnhandledTopic_Ignores(t *testing.T) {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalOutput)

	mailer := &email.MockMailer{}
	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"oncaller": "test"}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.ServiceUpTopic, mailer)

	assert.False(t, mailer.SendCalled, "Mailer should NOT be called for wrong topic")
	assert.True(t, msg.Acked, "Message should be ACKed")
}
