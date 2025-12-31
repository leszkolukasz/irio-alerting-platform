package main

import (
	"alerting-plafform/incident-manager/internal"
	redis_keys "alerting-plafform/incident-manager/redis"
	"alerting-platform/common/config"
	"alerting-platform/common/db"
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/redis/go-redis/v9"

	pubsub_common "alerting-platform/common/pubsub"
)

func main() {
	config.Intro("Incident Manager")

	ctx := context.Background()

	psClient := pubsub_common.Init(ctx)
	defer psClient.Close()

	managerState := internal.NewManagerState(ctx, psClient)

	var wg sync.WaitGroup

	StartPubSubListener(ctx, &wg, psClient, managerState)
	StartIncidentManager(ctx, managerState)

	log.Println("[INFO] Incident Manager service is running...")

	wg.Wait()

}

func StartPubSubListener(ctx context.Context, wg *sync.WaitGroup, psClient *pubsub.Client, managerState *internal.ManagerState) {
	subscriptions := map[string]string{
		"incident-manager-service-up":            pubsub_common.ServiceUpTopic,
		"incident-manager-service-down":          pubsub_common.ServiceDownTopic,
		"incident-manager-service-created":       pubsub_common.ServiceCreatedTopic,
		"incident-manager-service-removed":       pubsub_common.ServiceRemovedTopic,
		"incident-manager-service-modified":      pubsub_common.ServiceModifiedTopic,
		"incident-manager-oncaller-acknowledged": pubsub_common.OncallerAcknowledgedTopic,
	}

	pubsub_common.CreateSubscriptionsAndTopics(psClient, subscriptions, []string{pubsub_common.NotifyOncallerTopic})
	pubsub_common.SetupSubscriptionListeners(ctx, psClient, subscriptions, wg, func(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
		managerState.HandleMessage(ctx, msg, eventType)
	})

	log.Println("[INFO] Pub/Sub listener started")
}

func StartIncidentManager(ctx context.Context, managerState *internal.ManagerState) {
	redisClient := db.GetRedisClient()

	go func() {
		for {
			oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
			now := time.Now().UTC().Unix()

			expiredDeadlines, err := redisClient.ZRangeByScore(ctx, oncallerDeadlineSetKey, &redis.ZRangeBy{
				Min: "-inf",
				Max: strconv.FormatInt(now, 10),
			}).Result()

			log.Printf("[DEBUG] Got %d expired deadlines", len(expiredDeadlines))

			if err != nil {
				log.Printf("[ERROR] Failed to fetch expired incidents from Redis: %v", err)
				time.Sleep(time.Second)
				continue
			}

			for _, serviceID := range expiredDeadlines {
				serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
				if err != nil {
					log.Printf("[ERROR] Invalid service ID %s: %v", serviceID, err)
					continue
				}

				go func() {
					err := managerState.HandleExpiredDeadline(ctx, serviceIDInt)
					if err != nil {
						log.Printf("[ERROR] Failed to handle expired deadline for service %d: %v", serviceIDInt, err)
					}
				}()
			}

			time.Sleep(15 * time.Second)
		}
	}()

}
