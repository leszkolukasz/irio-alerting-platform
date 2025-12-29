package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	db "logger/db"
	dto "logger/dto"
)

type Repository interface {
	SaveLog(context.Context, db.IncidentLog) error
	SaveMetric(context.Context, db.MetricLog) error
}

func HandleMessage(
	ctx context.Context,
	msg PubSubMessage,
	eventType string,
	repo Repository,
) {
	var payload dto.PubSubPayload

	if err := json.Unmarshal(msg.GetData(), &payload); err != nil {
		log.Printf("[CRITICAL] Error unmarshalling JSON from %s: %v. Dropping message.", eventType, err)
		msg.Ack()
		return
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

	var err error
	if eventType == "ServiceUp" || eventType == "ServiceDown" {
		err = repo.SaveMetric(ctx, db.MetricLog{
			ServiceID: payload.ServiceID,
			Timestamp: eventTime,
			Type:      eventType,
		})
	} else {
		err = repo.SaveLog(ctx, db.IncidentLog{
			IncidentID:   payload.IncidentID,
			OnCallerData: payload.OnCallerData,
			Timestamp:    eventTime,
			Type:         eventType,
		})
	}

	if err != nil {
		msg.Nack()
	} else {
		msg.Ack()
	}
}
