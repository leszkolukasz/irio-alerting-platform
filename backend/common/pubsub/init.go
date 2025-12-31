package pubsub

import (
	"alerting-platform/common/config"
	"context"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

var (
	psClient *pubsub.Client
)

func Init(ctx context.Context) *pubsub.Client {
	cfg := config.GetConfig()
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)

	if err != nil {
		log.Fatalf("[FATAL] Failed to init Pub/Sub: %v", err)
	}

	psClient = client
	return psClient
}

func GetClient() *pubsub.Client {
	return psClient
}

func CreateSubscriptionsAndTopics(psClient *pubsub.Client, subscriptions map[string]string, topics []string) {
	if config.GetConfig().Env != config.DEV {
		return
	}

	for subID, topicID := range subscriptions {
		topic := psClient.Topic(topicID)
		exists, err := topic.Exists(context.Background())

		if err != nil {
			log.Fatalf("[FATAL] Failed to check if topic %s exists: %v", topicID, err)
		}

		if !exists {
			topic, err = psClient.CreateTopic(context.Background(), topicID)
			if err != nil {
				log.Fatalf("[FATAL] Failed to create topic %s: %v", topicID, err)
			}

			log.Printf("[INFO] Created topic: %s", topicID)
		}

		sub := psClient.Subscription(subID)
		exists, err = sub.Exists(context.Background())
		if err != nil {
			log.Fatalf("[FATAL] Failed to check if subscription %s exists: %v", subID, err)
		}

		if !exists {
			_, err = psClient.CreateSubscription(context.Background(), subID, pubsub.SubscriptionConfig{
				Topic:                 topic,
				EnableMessageOrdering: true,
			})
			if err != nil {
				log.Fatalf("[FATAL] Failed to create subscription %s: %v", subID, err)
			}

			log.Printf("[INFO] Created subscription: %s", subID)
		}
	}

	for _, topicID := range topics {
		topic := psClient.Topic(topicID)
		exists, err := topic.Exists(context.Background())

		if err != nil {
			log.Fatalf("[FATAL] Failed to check if topic %s exists: %v", topicID, err)
		}

		if !exists {
			_, err = psClient.CreateTopic(context.Background(), topicID)
			if err != nil {
				log.Fatalf("[FATAL] Failed to create topic %s: %v", topicID, err)
			}

			log.Printf("[INFO] Created topic: %s", topicID)
		}
	}
}

func SetupSubscriptionListeners(ctx context.Context, psClient *pubsub.Client, subscriptions map[string]string, wg *sync.WaitGroup,
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
				log.Printf("[ERROR] Receive error on %s: %v", sid, err)
			}
		}(subID, eventType)
	}
}

func HealthCheck(psClient *pubsub.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topics := psClient.Topics(ctx)
	_, err := topics.Next()
	if err != nil && err != iterator.Done {
		return false
	}
	return true
}
