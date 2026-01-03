package main

import (
	"alerting-platform/common/config"
	pubsub_common "alerting-platform/common/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
)

func checkHealth(url string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Request failed for %s: %v", url, err)
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func main() {
	config.Intro("Worker")
	ctx := context.Background()

	pubsubClient := pubsub_common.Init(ctx)
	defer pubsubClient.Close()

	log.Println("Worker service is running and connected to Pub/Sub ...")

	topicName := "execute-health-check"
	subName := "worker-execute-health-check"

	subscriptions := map[string]string{
		subName: topicName,
	}

	pubsub_common.CreateSubscriptionsAndTopics(pubsubClient, subscriptions, nil)

	sub := pubsubClient.Subscription(subName)

	// this is the limit of simultaneous tasks that the worker
	// can process:
	sub.ReceiveSettings.MaxOutstandingMessages = 10

	log.Printf("Worker listening on subscription %s ...", subName)

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var task pubsub_common.MonitoringTask
		if err := json.Unmarshal(msg.Data, &task); err != nil {
			log.Printf("Error decoding message: %v", err)
			msg.Ack()
			return
		}

		log.Printf("[Worker] Recived task: Check %s serviceId: %d", task.URL, task.ServiceID)

		isServiceUp := checkHealth(task.URL)
		resultTopic := pubsub_common.ServiceDownTopic
		if isServiceUp {
			resultTopic = pubsub_common.ServiceUpTopic
		}

		payload := pubsub_common.PubSubPayload{
			ServiceID: task.ServiceID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		err := pubsub_common.SendMessage(ctx, pubsubClient, resultTopic, payload, fmt.Sprintf("%d", task.ServiceID))
		if err != nil {
			log.Printf("Failed to publish result to %s: %v", resultTopic, err)
			msg.Nack()
		} else {
			log.Printf("[Worker] Successfully processed task for service %d. Result reported to %s", task.ServiceID, resultTopic)
			msg.Ack()
		}
	})

	if err != nil {
		log.Fatalf("Receiver failed: %v", err)
	}

}
