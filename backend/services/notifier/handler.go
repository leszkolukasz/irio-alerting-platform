package main

import (
	"context"
	"log"
	"notifier/email"

	"alerting-platform/common/pubsub"
)

var EventTypeToStatus = map[string]string{
	pubsub.NotifyOncallerTopic: "NOTIFY",
}

func HandleMessage(
	ctx context.Context,
	msg pubsub.PubSubMessage,
	eventType string,
	mailer *email.Mailer,
) {
	payload, _, err := pubsub.ExtractPayload(msg)
	if err != nil {
		log.Printf("[CRITICAL] Error extracting payload for topic %s: %v. Dropping message.", eventType, err)
		msg.Ack()
		return
	}

	switch eventType {
	case pubsub.NotifyOncallerTopic:
		if sendErr := mailer.SendNotification(payload.OnCaller, payload.IncidentID, payload.ServiceID); sendErr != nil {
			log.Printf("[ERROR] Failed to notify oncaller %s: %v", payload.OnCaller, sendErr)
		}
	default:
		log.Printf("[WARNING] Unhandled event type: %s", eventType)
	}

	msg.Ack()
}
