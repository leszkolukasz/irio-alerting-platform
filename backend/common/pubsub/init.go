package pubsub

import (
	"alerting-platform/common/config"
	"context"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
)

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
			_, err = psClient.CreateSubscription(context.Background(), subID, pubsub.SubscriptionConfig{
				Topic: topic,
			})
			if err != nil {
				log.Fatalf("Failed to create subscription %s: %v", subID, err)
			}

			log.Printf("Created subscription: %s", subID)
		}
	}
}

func SetupSubscriptions(ctx context.Context, psClient *pubsub.Client, subscriptions map[string]string, wg *sync.WaitGroup,
	handler_func func(context.Context, PubSubMessage, string)) {
	for subID, eventType := range subscriptions {
		wg.Add(1)

		go func(sid, eType string) {
			defer wg.Done()

			sub := psClient.Subscription(sid)
			sub.ReceiveSettings.MaxOutstandingMessages = -1 // TODO: How to tune this?

			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				adapter := &PubSubMessageAdapter{Msg: msg}
				handler_func(ctx, adapter, eType)
			})

			if err != nil {
				log.Printf("Receive error on %s: %v", sid, err)
			}
		}(subID, eventType)
	}
}
