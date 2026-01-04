package main

// the only purpose of this file is to setup topics in the message broker
import (
	"alerting-platform/common/config"
	pubsub_common "alerting-platform/common/pubsub"
	"context"
	"log"
)

func main() {
	config.Intro("Script for creating topics ...")
	ctx := context.Background()

	psClient := pubsub_common.Init(ctx)
	defer psClient.Close()

	subscriptions := map[string]string{
		"worker-execute-health-check": "execute-health-check",
	}

	outputTopics := []string{
		"service-up",
		"service-down",
	}

	log.Println("Creating missing subscriptions and topics ...")
	pubsub_common.CreateSubscriptionsAndTopics(psClient, subscriptions, outputTopics)

	log.Println("Infrastructure is ready.")
}
