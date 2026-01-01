package firestore

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) SaveLog(ctx context.Context, log IncidentLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockLogRepository) SaveMetric(ctx context.Context, metric MetricLog) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockLogRepository) GetIncidentsByService(ctx context.Context, serviceID uint) ([]IncidentLog, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]IncidentLog), args.Error(1)
}

func (m *MockLogRepository) GetMetricsByServiceAndAfterTime(ctx context.Context, serviceID uint, afterTime time.Time) ([]MetricLog, error) {
	args := m.Called(ctx, serviceID, afterTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]MetricLog), args.Error(1)
}
