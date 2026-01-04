package pubsub

import (
	"alerting-platform/api/db"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) SendServiceCreatedMessage(ctx context.Context, service db.MonitoredService) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockPubSubService) SendServiceUpdatedMessage(ctx context.Context, service db.MonitoredService) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockPubSubService) SendServiceDeletedMessage(ctx context.Context, serviceID uint64) error {
	args := m.Called(ctx, serviceID)
	return args.Error(0)
}

func (m *MockPubSubService) SendOncallerAcknowledgedMessage(ctx context.Context, incidentID string, onCaller string) error {
	args := m.Called(ctx, incidentID, onCaller)
	return args.Error(0)
}
