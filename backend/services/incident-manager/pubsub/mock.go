package pubsub

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) SendIncidentStartMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	args := m.Called(ctx, incidentID, serviceID, timestamp)
	return args.Error(0)
}

func (m *MockPubSubService) SendAcknowledgeTimeoutMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	args := m.Called(ctx, incidentID, serviceID, oncaller, timestamp)
	return args.Error(0)
}

func (m *MockPubSubService) SendNotifyOncallerMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	args := m.Called(ctx, incidentID, serviceID, oncaller, timestamp)
	return args.Error(0)
}

func (m *MockPubSubService) SendIncidentUnresolvedMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	args := m.Called(ctx, incidentID, serviceID, timestamp)
	return args.Error(0)
}

func (m *MockPubSubService) SendIncidentResolvedMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	args := m.Called(ctx, incidentID, serviceID, oncaller, timestamp)
	return args.Error(0)
}
