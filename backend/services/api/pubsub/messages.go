package pubsub

import (
	"alerting-platform/api/db"
	"alerting-platform/common/pubsub"
	"context"
	"fmt"
	"time"
)

func SendServiceCreatedMessage(ctx context.Context, service db.MonitoredService) error {
	psClient := pubsub.GetClient()

	oncallers := []string{service.FirstOncallerEmail}

	if service.SecondOncallerEmail != nil {
		oncallers = append(oncallers, *service.SecondOncallerEmail)
	}

	payload := pubsub.PubSubPayload{
		ServiceID: uint64(service.ID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: pubsub.PubSubPayloadData{
			AllowedResponseTime: service.AllowedResponseTime,
			AlertWindow:         service.AlertWindow,
			HealthCheckInterval: service.HealthCheckInterval,
			Oncallers:           oncallers,
		},
	}

	return pubsub.SendMessage(ctx, psClient, pubsub.ServiceCreatedTopic, payload, fmt.Sprintf("%d", service.ID))
}

func SendServiceUpdatedMessage(ctx context.Context, service db.MonitoredService) error {
	psClient := pubsub.GetClient()

	oncallers := []string{service.FirstOncallerEmail}

	if service.SecondOncallerEmail != nil {
		oncallers = append(oncallers, *service.SecondOncallerEmail)
	}

	payload := pubsub.PubSubPayload{
		ServiceID: uint64(service.ID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: pubsub.PubSubPayloadData{
			AllowedResponseTime: service.AllowedResponseTime,
			AlertWindow:         service.AlertWindow,
			HealthCheckInterval: service.HealthCheckInterval,
			Oncallers:           oncallers,
		},
	}

	return pubsub.SendMessage(ctx, psClient, pubsub.ServiceModifiedTopic, payload, fmt.Sprintf("%d", service.ID))
}

func SendServiceDeletedMessage(ctx context.Context, serviceID uint64) error {
	psClient := pubsub.GetClient()

	payload := pubsub.PubSubPayload{
		ServiceID: serviceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return pubsub.SendMessage(ctx, psClient, pubsub.ServiceRemovedTopic, payload, fmt.Sprintf("%d", serviceID))
}
