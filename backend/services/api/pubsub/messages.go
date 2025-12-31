package pubsub

import (
	"alerting-platform/api/db"
	pubsub_common "alerting-platform/common/pubsub"
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type PubSubServiceI interface {
	SendServiceCreatedMessage(ctx context.Context, service db.MonitoredService) error
	SendServiceUpdatedMessage(ctx context.Context, service db.MonitoredService) error
	SendServiceDeletedMessage(ctx context.Context, serviceID uint64) error
}

type PubSubService struct {
	client *pubsub.Client
}

func NewPubSubService(client *pubsub.Client) *PubSubService {
	return &PubSubService{client: client}
}

func (s *PubSubService) SendServiceCreatedMessage(ctx context.Context, service db.MonitoredService) error {
	oncallers := []string{service.FirstOncallerEmail}

	if service.SecondOncallerEmail != nil {
		oncallers = append(oncallers, *service.SecondOncallerEmail)
	}

	payload := pubsub_common.PubSubPayload{
		ServiceID: uint64(service.ID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: pubsub_common.PubSubPayloadData{
			AllowedResponseTime: service.AllowedResponseTime,
			AlertWindow:         service.AlertWindow,
			HealthCheckInterval: service.HealthCheckInterval,
			Oncallers:           oncallers,
		},
	}

	return pubsub_common.SendMessage(ctx, s.client, pubsub_common.ServiceCreatedTopic, payload, fmt.Sprintf("%d", service.ID))
}

func (s *PubSubService) SendServiceUpdatedMessage(ctx context.Context, service db.MonitoredService) error {

	oncallers := []string{service.FirstOncallerEmail}

	if service.SecondOncallerEmail != nil {
		oncallers = append(oncallers, *service.SecondOncallerEmail)
	}

	payload := pubsub_common.PubSubPayload{
		ServiceID: uint64(service.ID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: pubsub_common.PubSubPayloadData{
			AllowedResponseTime: service.AllowedResponseTime,
			AlertWindow:         service.AlertWindow,
			HealthCheckInterval: service.HealthCheckInterval,
			Oncallers:           oncallers,
		},
	}

	return pubsub_common.SendMessage(ctx, s.client, pubsub_common.ServiceModifiedTopic, payload, fmt.Sprintf("%d", service.ID))
}

func (s *PubSubService) SendServiceDeletedMessage(ctx context.Context, serviceID uint64) error {
	payload := pubsub_common.PubSubPayload{
		ServiceID: serviceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return pubsub_common.SendMessage(ctx, s.client, pubsub_common.ServiceRemovedTopic, payload, fmt.Sprintf("%d", serviceID))
}
