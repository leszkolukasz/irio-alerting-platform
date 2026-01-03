package main

import (
	"alerting-platform/common/config"
	pubsub_common "alerting-platform/common/pubsub"
	"context"
	"log"
	email "notifier/email"
	"sync"
)

func main() {
	config.Intro("Notifier")

	ctx := context.Background()

	// Pub/Sub
	psClient := pubsub_common.Init(ctx)
	defer psClient.Close()

	// Mailer
	mailer, err := email.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize mailer: %v", err)
		return
	}

	subscriptions := map[string]string{
		"notifier-notify-oncaller": pubsub_common.NotifyOncallerTopic,
	}

	pubsub_common.CreateSubscriptionsAndTopics(psClient, subscriptions, []string{})

	var wg sync.WaitGroup
	pubsub_common.SetupSubscriptionListeners(ctx, psClient, subscriptions, &wg,
		func(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
			HandleMessage(ctx, msg, eventType, mailer)
		})

	log.Println("Notifier service started and listening to Pub/Sub subscriptions...")

	wg.Wait()
}
