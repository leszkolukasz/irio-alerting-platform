package main

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	db "alerting-platform/common/db/firestore"
	pubsub "alerting-platform/common/pubsub"
)

type mockRepo struct {
	saveLogCalled    bool
	saveMetricCalled bool

	lastIncident db.IncidentLog
	lastMetric   db.MetricLog

	err error
}

func (m *mockRepo) SaveLog(ctx context.Context, l db.IncidentLog) error {
	m.saveLogCalled = true
	m.lastIncident = l
	return m.err
}

func (m *mockRepo) SaveMetric(ctx context.Context, mlog db.MetricLog) error {
	m.saveMetricCalled = true
	m.lastMetric = mlog
	return m.err
}

func (m *mockRepo) GetIncidentsByService(ctx context.Context, serviceID uint) ([]db.IncidentLog, error) {
	return nil, nil
}

func (m *mockRepo) GetMetricsByServiceAndAfterTime(ctx context.Context, serviceID uint, afterTime time.Time) ([]db.MetricLog, error) {
	return nil, nil
}

func TestHandleMessage_Incident_Ack(t *testing.T) {
	repo := &mockRepo{}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"incident_id":"inc-1", "service_id": 1}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.IncidentStartTopic, repo)

	assert.True(t, repo.saveLogCalled, "SaveLog should be called")
	assert.False(t, repo.saveMetricCalled, "SaveMetric should NOT be called")

	assert.True(t, msg.Acked, "Message should be ACKed")
	assert.False(t, msg.Nacked, "Message should NOT be NACKed")
}

func TestHandleMessage_DB_NackOnError(t *testing.T) {
	repo := &mockRepo{
		err: errors.New("db error"),
	}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"service_id": 1}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.ServiceUpTopic, repo)

	assert.True(t, repo.saveMetricCalled, "SaveMetric should be called")
	assert.False(t, msg.Acked, "Message should NOT be ACKed")
	assert.True(t, msg.Nacked, "Message should be NACKed")
}

func TestHandleMessage_InvalidJSON_Ack(t *testing.T) {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalOutput)

	repo := &mockRepo{}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{invalid-json}`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.IncidentStartTopic, repo)

	assert.False(t, repo.saveLogCalled, "SaveLog should NOT be called")
	assert.False(t, repo.saveMetricCalled, "SaveMetric should NOT be called")

	assert.True(t, msg.Acked, "Invalid message should be ACKed (dropped)")
}

func TestHandleMessage_UsesPayloadTimestamp(t *testing.T) {
	repo := &mockRepo{}

	ts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)

	msg := &pubsub.FakeMessage{
		Data: []byte(`{
            "incident_id": "inc-2",
            "service_id": 2,
            "timestamp": "` + ts + `"
        }`),
		PublishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, pubsub.IncidentStartTopic, repo)

	assert.True(t, repo.saveLogCalled)
	assert.WithinDuration(t, time.Now().Add(-time.Hour), repo.lastIncident.Timestamp, time.Second)
}

func TestHandleMessage_UsesNowIfZero(t *testing.T) {
	repo := &mockRepo{}

	msg := &pubsub.FakeMessage{
		Data:        []byte(`{"incident_id":"inc-3", "service_id": 3}`),
		PublishTime: time.Time{},
	}

	before := time.Now().UTC()

	HandleMessage(context.Background(), msg, pubsub.IncidentStartTopic, repo)

	after := time.Now().UTC()

	assert.True(t, repo.saveLogCalled, "SaveLog should be called")
	assert.True(t, msg.Acked, "Message should be ACKed")
	assert.False(t, msg.Nacked, "Message should NOT be NACKed")

	assert.True(t,
		repo.lastIncident.Timestamp.After(before) && repo.lastIncident.Timestamp.Before(after.Add(time.Second)),
		"Timestamp should be set to current time if zero",
	)
}
