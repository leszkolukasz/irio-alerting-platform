package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	db "alerting-platform/common/db/firestore"
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

type fakeMessage struct {
	data        []byte
	publishTime time.Time
	acked       bool
	nacked      bool
}

func (m *fakeMessage) Ack()                      { m.acked = true }
func (m *fakeMessage) Nack()                     { m.nacked = true }
func (m *fakeMessage) GetData() []byte           { return m.data }
func (m *fakeMessage) GetPublishTime() time.Time { return m.publishTime }

func TestHandleMessage_Incident_Ack(t *testing.T) {
	repo := &mockRepo{}

	msg := &fakeMessage{
		data:        []byte(`{"incident_id":"inc-1"}`),
		publishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, "IncidentStart", repo)

	assert.True(t, repo.saveLogCalled, "SaveLog should be called")
	assert.False(t, repo.saveMetricCalled, "SaveMetric should NOT be called")

	assert.True(t, msg.acked, "Message should be ACKed")
	assert.False(t, msg.nacked, "Message should NOT be NACKed")
}

func TestHandleMessage_DB_NackOnError(t *testing.T) {
	repo := &mockRepo{
		err: errors.New("db error"),
	}

	msg := &fakeMessage{
		data:        []byte(`{"service_id":"svc-1"}`),
		publishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, "ServiceUp", repo)

	assert.True(t, repo.saveMetricCalled, "SaveMetric should be called")
	assert.False(t, msg.acked, "Message should NOT be ACKed")
	assert.True(t, msg.nacked, "Message should be NACKed")
}

func TestHandleMessage_InvalidJSON_Ack(t *testing.T) {
	repo := &mockRepo{}

	msg := &fakeMessage{
		data:        []byte(`{invalid-json}`),
		publishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, "IncidentStart", repo)

	assert.False(t, repo.saveLogCalled, "SaveLog should NOT be called")
	assert.False(t, repo.saveMetricCalled, "SaveMetric should NOT be called")

	assert.True(t, msg.acked, "Invalid message should be ACKed (dropped)")
}

func TestHandleMessage_UsesPayloadTimestamp(t *testing.T) {
	repo := &mockRepo{}

	ts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)

	msg := &fakeMessage{
		data: []byte(`{
			"incident_id": "inc-2",
			"timestamp": "` + ts + `"
		}`),
		publishTime: time.Now().UTC(),
	}

	HandleMessage(context.Background(), msg, "IncidentStart", repo)

	assert.True(t, repo.saveLogCalled)
	assert.WithinDuration(t, time.Now().Add(-time.Hour), repo.lastIncident.Timestamp, time.Second)
}

func TestHandleMessage_UsesNowIfZero(t *testing.T) {
	repo := &mockRepo{}

	msg := &fakeMessage{
		data:        []byte(`{"incident_id":"inc-3"}`),
		publishTime: time.Time{},
	}

	before := time.Now().UTC()

	HandleMessage(context.Background(), msg, "IncidentStart", repo)

	after := time.Now().UTC()

	assert.True(t, repo.saveLogCalled, "SaveLog should be called")
	assert.True(t, msg.acked, "Message should be ACKed")
	assert.False(t, msg.nacked, "Message should NOT be NACKed")

	assert.True(t,
		repo.lastIncident.Timestamp.After(before) && repo.lastIncident.Timestamp.Before(after.Add(time.Second)),
		"Timestamp should be set to current time if zero",
	)
}
