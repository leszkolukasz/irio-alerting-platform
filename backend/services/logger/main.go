package main

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"

	"alerting-platform/common/config"
	db "logger/db"
)

func main() {
	ctx := context.Background()
	cfg := config.GetConfig()

	if cfg.ProjectID == "" {
		log.Fatal("PROJECT_ID is required")
	}

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
		"logger-incident-start":      "IncidentStart",
		"logger-incident-resolved":   "IncidentResolved",
		"logger-incident-timeout":    "IncidentAcknowledgeTimeout",
		"logger-incident-unresolved": "IncidentUnresolved",
		"logger-service-up":          "ServiceUp",
		"logger-service-down":        "ServiceDown",
	}

	Init(psClient, subscriptions)

	var wg sync.WaitGroup

	for subID, eventType := range subscriptions {
		wg.Add(1)

		go func(sid, eType string) {
			defer wg.Done()

			sub := psClient.Subscription(sid)
			sub.ReceiveSettings.MaxOutstandingMessages = 10

			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				adapter := &PubSubMessageAdapter{msg: msg}
				HandleMessage(ctx, adapter, eType, repo)
			})

			if err != nil {
				log.Printf("Receive error on %s: %v", sid, err)
			}
		}(subID, eventType)
	}
	log.Println("Logger service started and listening to Pub/Sub subscriptions...")

	wg.Wait()
}

func Init(psClient *pubsub.Client, subscriptions map[string]string) {
	if config.GetConfig().Env != config.DEV {
		return
	}

	for subID, topicID := range subscriptions {
		topic := psClient.Topic(topicID)
		exists, err := topic.Exists(context.Background())

		if err != nil {
			log.Fatalf("Failed to check if topic %s exists: %v", topicID, err)
		}

		if !exists {
			topic, err = psClient.CreateTopic(context.Background(), topicID)
			if err != nil {
				log.Fatalf("Failed to create topic %s: %v", topicID, err)
			}

			log.Printf("Created topic: %s", topicID)
		}

		sub := psClient.Subscription(subID)
		exists, err = sub.Exists(context.Background())
		if err != nil {
			log.Fatalf("Failed to check if subscription %s exists: %v", subID, err)
		}

		if !exists {
			sub, err = psClient.CreateSubscription(context.Background(), subID, pubsub.SubscriptionConfig{
				Topic: topic,
			})
			if err != nil {
				log.Fatalf("Failed to create subscription %s: %v", subID, err)
			}

			log.Printf("Created subscription: %s", subID)
		}
	}
}
