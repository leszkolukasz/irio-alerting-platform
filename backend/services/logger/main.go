package main

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"

	"alerting-platform/common/config"
	pubsub_common "alerting-platform/common/pubsub"
	db "logger/db"
)

func main() {
	ctx := context.Background()
	cfg := config.GetConfig()

	// Firestore
	repo, err := db.NewLogRepository(ctx, cfg.ProjectID, cfg.FirestoreDB)
	if err != nil {
		log.Fatalf("Failed to init Firestore: %v", err)
	}
	defer repo.Close()

	// Pub/Sub
	psClient, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to init Pub/Sub: %v", err)
	}
	defer psClient.Close()

	subscriptions := map[string]string{
		"logger-incident-start":      pubsub_common.IncidentStartTopic,
		"logger-incident-resolved":   pubsub_common.IncidentResolvedTopic,
		"logger-incident-timeout":    pubsub_common.IncidentAcknowledgeTimeoutTopic,
		"logger-incident-unresolved": pubsub_common.IncidentUnresolvedTopic,
		"logger-service-up":          pubsub_common.ServiceUpTopic,
		"logger-service-down":        pubsub_common.ServiceDownTopic,
	}

	pubsub_common.Init(psClient, subscriptions)

	var wg sync.WaitGroup
	pubsub_common.SetupSubscriptions(ctx, psClient, subscriptions, &wg,
		func(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
			HandleMessage(ctx, msg, eventType, repo)
		})

	log.Println("Logger service started and listening to Pub/Sub subscriptions...")

	wg.Wait()
}
