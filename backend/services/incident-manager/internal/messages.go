package internal

import (
	pubsub_common "alerting-platform/common/pubsub"
	"context"
	"log"
	"time"
)

func (managerState *ManagerState) SendIncidentStartMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentStart message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendMessage(ctx, managerState.psClient, pubsub_common.IncidentStartTopic, payload)
}

func (managerState *ManagerState) SendAcknowledgeTimeoutMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentAcknowledgeTimeout message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.OnCaller = oncaller
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendMessage(ctx, managerState.psClient, pubsub_common.IncidentAcknowledgeTimeoutTopic, payload)
}

func (managerState *ManagerState) SendIncidentUnresolvedMessage(ctx context.Context, incidentID string, serviceID uint64, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentUnresolved message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendMessage(ctx, managerState.psClient, pubsub_common.IncidentUnresolvedTopic, payload)
}

func (managerState *ManagerState) SendIncidentResolvedMessage(ctx context.Context, incidentID string, serviceID uint64, oncaller string, timestamp time.Time) error {
	var payload pubsub_common.PubSubPayload

	log.Printf("[DEBUG] Sending IncidentResolved message")

	payload.IncidentID = incidentID
	payload.ServiceID = serviceID
	payload.OnCaller = oncaller
	payload.Timestamp = timestamp.Format(time.RFC3339)

	return pubsub_common.SendMessage(ctx, managerState.psClient, pubsub_common.IncidentResolvedTopic, payload)
}
