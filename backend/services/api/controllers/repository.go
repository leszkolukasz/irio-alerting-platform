package controllers

import (
	"alerting-platform/api/db"
	"context"

	"gorm.io/gorm"
)

type RepositoryI interface {
	GetServiceByName(ctx context.Context, name string) (*db.MonitoredService, error)
	GetServicesForUser(ctx context.Context, userID uint64) ([]db.MonitoredService, error)
	GetServiceByIDAndUserID(ctx context.Context, serviceID uint64, userID uint64) (*db.MonitoredService, error)
	CreateService(ctx context.Context, service *db.MonitoredService) error
	SaveService(ctx context.Context, service *db.MonitoredService)
	DeleteServiceForUser(ctx context.Context, serviceID uint64, userID uint64) (int, error)
	CreateUser(ctx context.Context, user *db.User) error
}

type Repository struct {
	conn *gorm.DB
}

func NewRepository(conn *gorm.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) GetServiceByName(ctx context.Context, name string) (*db.MonitoredService, error) {
	service, err := gorm.G[db.MonitoredService](r.conn).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *Repository) GetServicesForUser(ctx context.Context, userID uint64) ([]db.MonitoredService, error) {
	services, err := gorm.G[db.MonitoredService](r.conn).Where("user_id = ?", userID).Find(ctx)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (r *Repository) GetServiceByIDAndUserID(ctx context.Context, serviceID uint64, userID uint64) (*db.MonitoredService, error) {
	service, err := gorm.G[db.MonitoredService](r.conn).Where("id = ? AND user_id = ?", serviceID, userID).First(ctx)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *Repository) CreateService(ctx context.Context, service *db.MonitoredService) error {
	return gorm.G[db.MonitoredService](r.conn).Create(ctx, service)
}

func (r *Repository) SaveService(ctx context.Context, service *db.MonitoredService) {
	r.conn.Save(service)
}

func (r *Repository) DeleteServiceForUser(ctx context.Context, serviceID uint64, userID uint64) (int, error) {
	return gorm.G[db.MonitoredService](r.conn).Where("id = ? AND user_id = ?", serviceID, userID).Delete(ctx)
}

func (r *Repository) CreateUser(ctx context.Context, user *db.User) error {
	return r.conn.Create(user).Error
}
