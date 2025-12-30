package main

import (
	"context"
	"fmt"
	"log"

	"alerting-platform/common/pubsub"
	db "logger/db"
)

type Repository interface {
	SaveLog(context.Context, db.IncidentLog) error
	SaveMetric(context.Context, db.MetricLog) error
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

	if eventType == "ServiceUp" || eventType == "ServiceDown" {
		err = repo.SaveMetric(ctx, db.MetricLog{
			ServiceID: fmt.Sprintf("%d", payload.ServiceID),
			Timestamp: *eventTime,
			Type:      eventType,
		})
	} else {
		err = repo.SaveLog(ctx, db.IncidentLog{
			IncidentID: payload.IncidentID,
			Oncaller:   payload.OnCaller,
			Timestamp:  *eventTime,
			Type:       eventType,
		})
	}

	if err != nil {
		msg.Nack()
	} else {
		msg.Ack()
	}
}
