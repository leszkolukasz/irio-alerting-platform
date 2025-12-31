package main

import (
	"context"
	"fmt"
	"log"

	firestore "alerting-platform/common/db/firestore"
	"alerting-platform/common/pubsub"
)

type Repository interface {
	SaveLog(context.Context, firestore.IncidentLog) error
	SaveMetric(context.Context, firestore.MetricLog) error
}

var EventTypeToStatus = map[string]string{
	pubsub.ServiceUpTopic:                  "UP",
	pubsub.ServiceDownTopic:                "DOWN",
	pubsub.IncidentStartTopic:              "START",
	pubsub.IncidentResolvedTopic:           "RESOLVED",
	pubsub.IncidentAcknowledgeTimeoutTopic: "TIMEOUT",
	pubsub.IncidentUnresolvedTopic:         "UNRESOLVED",
	pubsub.NotifyOncallerTopic:             "NOTIFIED",
}

func HandleMessage(
	ctx context.Context,
	msg pubsub.PubSubMessage,
	eventType string,
	repo Repository,
) {
	payload, eventTime, err := pubsub.ExtractPayload(msg)
	if err != nil {
		log.Printf("[CRITICAL] Error extracting payload for topic %s: %v. Dropping message.", eventType, err)
		msg.Ack()
		return
	}

	switch eventType {
	case pubsub.ServiceUpTopic, pubsub.ServiceDownTopic:
		err = repo.SaveMetric(ctx, firestore.MetricLog{
			ServiceID: fmt.Sprintf("%d", payload.ServiceID),
			Timestamp: *eventTime,
			Type:      EventTypeToStatus[eventType],
		})
	case pubsub.IncidentStartTopic, pubsub.IncidentResolvedTopic, pubsub.IncidentAcknowledgeTimeoutTopic,
		pubsub.IncidentUnresolvedTopic, pubsub.NotifyOncallerTopic:
		err = repo.SaveLog(ctx, firestore.IncidentLog{
			IncidentID: payload.IncidentID,
			ServiceID:  int64(payload.ServiceID),
			Oncaller:   payload.OnCaller,
			Timestamp:  *eventTime,
			Type:       EventTypeToStatus[eventType],
		})
	default:
		log.Printf("[WARNING] Unhandled event type: %s", eventType)
	}

	if err != nil {
		msg.Nack()
	} else {
		msg.Ack()
	}
}
