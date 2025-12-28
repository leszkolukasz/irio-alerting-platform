package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"

	"alerting-platform/common/config"
	db "logger/db"
	dto "logger/dto"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg := config.GetConfig()

	if cfg.ProjectID == "" {
		log.Fatal("PROJECT_ID is required")
	}

	log.Printf("Starting Logger Service in env: %s, Project: %s", cfg.Env, cfg.ProjectID)

	// Initialize Firestore
	repo, err := db.NewLogRepository(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to init Firestore: %v", err)
	}
	defer repo.Close()

	// Initialize Pub/Sub
	psClient, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to init Pub/Sub: %v", err)
	}
	defer psClient.Close()

	// Mapping Subscription Name -> Event Type
	subscriptions := map[string]string{
		"logger-incident-start":      "IncidentStart",
		"logger-incident-resolved":   "IncidentResolved",
		"logger-incident-timeout":    "IncidentAcknowledgeTimeout",
		"logger-incident-unresolved": "IncidentUnresolved",
		"logger-service-up":          "ServiceUp",
		"logger-service-down":        "ServiceDown",
	}

	var wg sync.WaitGroup

	for subID, eventType := range subscriptions {
		wg.Add(1)
		go func(sid, eType string) {
			defer wg.Done()
			sub := psClient.Subscription(sid)
			sub.ReceiveSettings.MaxOutstandingMessages = 10

			log.Printf("Listening on subscription: %s", sid)

			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				var payload dto.PubSubPayload

				if err := json.Unmarshal(msg.Data, &payload); err != nil {
					log.Printf("[CRITICAL] Error unmarshalling JSON from %s: %v. Dropping message.", sid, err)
					msg.Ack()
					return
				}

				eventTime := msg.PublishTime
				if payload.Timestamp != "" {
					if t, err := time.Parse(time.RFC3339, payload.Timestamp); err == nil {
						eventTime = t
					}
				}
				if eventTime.IsZero() {
					eventTime = time.Now()
				}

				var saveErr error
				if eType == "ServiceUp" || eType == "ServiceDown" {
					metricEntry := db.MetricLog{
						ServiceID: payload.ServiceID,
						Timestamp: eventTime,
						Type:      eType,
					}

					saveErr = repo.SaveMetric(ctx, metricEntry)
				} else {
					logEntry := db.IncidentLog{
						IncidentID:   payload.IncidentID,
						OnCallerData: payload.OnCallerData,
						Timestamp:    eventTime,
						Type:         eType,
					}

					saveErr = repo.SaveLog(ctx, logEntry)
				}

				if saveErr != nil {
					log.Printf("Error saving to DB (Type: %s): %v. Sending NACK.", eType, saveErr)
					msg.Nack()
				} else {
					msg.Ack()
				}
			})

			if err != nil {
				log.Printf("Receive error on %s: %v", sid, err)
			}
		}(subID, eventType)
	}

	log.Println("Logger Service is running...")
	wg.Wait()
}
