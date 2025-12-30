package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type PubSubPayload struct {
	IncidentID string `json:"incident_id,omitempty"`
	ServiceID  uint64 `json:"service_id,omitempty"`
	OnCaller   string `json:"oncaller,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
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
		eventTime = time.Now()
	}

	return &payload, &eventTime, nil
}

func SendMessage(ctx context.Context, psClient *pubsub.Client, topicID string, payload PubSubPayload) error {
	topic := psClient.Topic(topicID)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[Error] Failed to marshal payload: %w", err)
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	_, err = result.Get(ctx)

	if err != nil {
		return fmt.Errorf("[Error] Failed to publish message to topic %s: %w", topicID, err)
	}

	return nil
}
