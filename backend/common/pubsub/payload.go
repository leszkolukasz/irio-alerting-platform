package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type MonitoringTask struct {
	ServiceID uint64 `json:"service_id"`
	URL       string `json:"url"`
}

type PubSubPayloadData struct {
	AllowedResponseTime int      `json:"allowed_response_time,omitempty"`
	HealthCheckInterval int      `json:"health_check_interval,omitempty"`
	AlertWindow         int      `json:"alert_window,omitempty"`
	Oncallers           []string `json:"oncallers,omitempty"`
}

type PubSubPayload struct {
	IncidentID string            `json:"incident_id,omitempty"`
	ServiceID  uint64            `json:"service_id,omitempty"`
	OnCaller   string            `json:"oncaller,omitempty"`
	Timestamp  string            `json:"timestamp,omitempty"`
	Data       PubSubPayloadData `json:"data,omitempty"`
}

func ExtractPayload(msg PubSubMessage) (*PubSubPayload, *time.Time, error) {
	var payload PubSubPayload

	if err := json.Unmarshal(msg.GetData(), &payload); err != nil {
		return nil, nil, err
	}

	eventTime := msg.GetPublishTime()
	if payload.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, payload.Timestamp); err == nil {
			eventTime = t
		}
	}
	if eventTime.IsZero() {
		eventTime = time.Now().UTC()
	}

	return &payload, &eventTime, nil
}

func SendPayload(ctx context.Context, psClient *pubsub.Client, topicID string, payload PubSubPayload, orderingKey string) error {
	topic := psClient.Topic(topicID)
	topic.EnableMessageOrdering = true

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[Error] Failed to marshal payload: %w", err)
	}

	return SendMessage(ctx, psClient, topicID, data, orderingKey)
}

func SendMessage(ctx context.Context, psClient *pubsub.Client, topicID string, data []byte, orderingKey string) error {
	topic := psClient.Topic(topicID)
	topic.EnableMessageOrdering = true

	result := topic.Publish(ctx, &pubsub.Message{
		Data:        data,
		OrderingKey: orderingKey,
	})

	_, err := result.Get(ctx)

	if err != nil {
		return fmt.Errorf("[Error] Failed to publish message to topic %s: %w", topicID, err)
	}

	return nil
}
