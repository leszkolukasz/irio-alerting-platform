package db

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func NewLogRepository(ctx context.Context, projectID string) (*LogRepository, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &LogRepository{client: client}, nil
}

func (r *LogRepository) SaveLog(ctx context.Context, logEntry IncidentLog) error {
	_, _, err := r.client.Collection("incident_logs").Add(ctx, logEntry)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		return err
	}
	return nil
}

func (r *LogRepository) SaveMetric(ctx context.Context, metricEntry MetricLog) error {
	_, _, err := r.client.Collection("metric_logs").Add(ctx, metricEntry)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		return err
	}
	return nil
}

func (r *LogRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
