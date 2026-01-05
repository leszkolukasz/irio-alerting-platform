package pubsub

import (
	pubsub_common "alerting-platform/common/pubsub"

	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
)

type PubSubServiceI interface {
	SendIncidentStartMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error
	SendAcknowledgeTimeoutMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error
	SendNotifyOncallerMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error
	SendIncidentUnresolvedMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error
	SendIncidentResolvedMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error
}

type PubSubService struct {
	client *pubsub.Client
}

func NewPubSubService(client *pubsub.Client) *PubSubService {
	return &PubSubService{client: client}
}

func (ps *PubSubService) SendIncidentStartMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentStart message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendPayload(ctx, ps.client, pubsub_common.IncidentStartTopic, payload, incidentID)
}

func (ps *PubSubService) SendAcknowledgeTimeoutMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentAcknowledgeTimeout message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.OnCaller = oncaller
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendPayload(ctx, ps.client, pubsub_common.IncidentAcknowledgeTimeoutTopic, payload, incidentID)
}

func (ps *PubSubService) SendNotifyOncallerMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending NotifyOncaller message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.OnCaller = oncaller
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendPayload(ctx, ps.client, pubsub_common.NotifyOncallerTopic, payload, incidentID)
}

func (ps *PubSubService) SendIncidentUnresolvedMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentUnresolved message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendPayload(ctx, ps.client, pubsub_common.IncidentUnresolvedTopic, payload, incidentID)
}

func (ps *PubSubService) SendIncidentResolvedMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentResolved message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.OnCaller = oncaller
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendPayload(ctx, ps.client, pubsub_common.IncidentResolvedTopic, payload, incidentID)
}
