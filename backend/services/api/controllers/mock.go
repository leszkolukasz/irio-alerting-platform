package controllers

import (
	"alerting-platform/api/db"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *db.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetServiceByName(ctx context.Context, name string) (*db.MonitoredService, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.MonitoredService), args.Error(1)
}

func (m *MockRepository) GetServicesForUser(ctx context.Context, userID uint64) ([]db.MonitoredService, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.MonitoredService), args.Error(1)
}

func (m *MockRepository) GetServiceByIDAndUserID(ctx context.Context, serviceID uint64, userID uint64) (*db.MonitoredService, error) {
	args := m.Called(ctx, serviceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.MonitoredService), args.Error(1)
}

func (m *MockRepository) CreateService(ctx context.Context, service *db.MonitoredService) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockRepository) SaveService(ctx context.Context, service *db.MonitoredService) {
	m.Called(ctx, service)
}

func (m *MockRepository) DeleteServiceForUser(ctx context.Context, serviceID uint64, userID uint64) (int, error) {
	args := m.Called(ctx, serviceID, userID)
	return args.Int(0), args.Error(1)
}

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
