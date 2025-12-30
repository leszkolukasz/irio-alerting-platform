package main

import (
	"alerting-plafform/incident-manager/internal"
	redis_keys "alerting-plafform/incident-manager/redis"
	"alerting-platform/common/db"
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/redis/go-redis/v9"

	"alerting-platform/common/config"
	pubsub_common "alerting-platform/common/pubsub"
)

func main() {
	ctx := context.Background()
	cfg := config.GetConfig()

	psClient, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("[FATAL] Failed to init Pub/Sub: %v", err)
	}
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
		"incident-manager-service-up":   pubsub_common.ServiceUpTopic,
		"incident-manager-service-down": pubsub_common.ServiceDownTopic,
	}

	pubsub_common.Init(psClient, subscriptions)
	pubsub_common.SetupSubscriptions(ctx, psClient, subscriptions, wg, func(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
		managerState.HandleMessage(ctx, msg, eventType)
	})

	log.Println("[INFO] Pub/Sub listener started")
}

func StartIncidentManager(ctx context.Context, managerState *internal.ManagerState) {
	redisClient := db.GetRedisClient()

	go func() {
		for {
			oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
			now := time.Now().Unix()

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
