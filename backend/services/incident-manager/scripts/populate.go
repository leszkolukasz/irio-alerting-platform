package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	ProjectID    = "test-project"
	EmulatorHost = "http://localhost:8085"
	ServiceID    = 123
	HoursBack    = 1
	Interval     = 1 * time.Second
)

type LogPayload struct {
	ServiceID int    `json:"service_id"`
	Timestamp string `json:"timestamp"`
}

type PubSubMessage struct {
	Messages []PubSubMessageItem `json:"messages"`
}

type PubSubMessageItem struct {
	Data string `json:"data"`
}

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(HoursBack) * time.Hour)

	count := 0
	for current := startTime; current.Before(endTime); current = current.Add(Interval) {
		status := "UP"
		if rand.Float32() < 0.1 {
			status = "DOWN"
		}

		payload := LogPayload{
			ServiceID: ServiceID,
			Timestamp: current.Format(time.RFC3339),
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Błąd JSON: %v\n", err)
			continue
		}

		dataBase64 := base64.StdEncoding.EncodeToString(payloadBytes)

		pubSubBody := PubSubMessage{
			Messages: []PubSubMessageItem{
				{Data: dataBase64},
			},
		}
		reqBody, _ := json.Marshal(pubSubBody)

		topicName := ""
		if status == "UP" {
			topicName = "service-up"
		} else {
			topicName = "service-down"
		}

		url := fmt.Sprintf("%s/v1/projects/%s/topics/%s:publish", EmulatorHost, ProjectID, topicName)

		resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Printf("Błąd sieci: %v\n", err)
			continue
		}
		resp.Body.Close()

		if count%100 == 0 {
			fmt.Printf("Wysłano log z: %s (Status: %s)\n", current.Format("15:04"), status)
		}
		count++
	}

	fmt.Printf("\nZakończono! Wysłano %d logów.\n", count)
}
