package firestore

import (
	"alerting-platform/common/config"
	"context"
	"log"
	"sync"

	"cloud.google.com/go/firestore"
)

var (
	repo *LogRepository
	once sync.Once
)

const (
	IncidentLogsCollection = "incident_logs"
	MetricLogsCollection   = "metric_logs"
)

func GetLogRepository(ctx context.Context) *LogRepository {
	once.Do(func() {
		cfg := config.GetConfig()
		rep, err := NewLogRepository(ctx, cfg.ProjectID, cfg.FirestoreDB)
		if err != nil {
			log.Fatalf("Failed to initialize LogRepository: %v", err)
		}
		repo = rep
	})

	return repo
}

func NewLogRepository(ctx context.Context, projectID string, firestoreDB string) (*LogRepository, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, firestoreDB)
	if err != nil {
		return nil, err
	}
	return &LogRepository{client: client}, nil
}

func (r *LogRepository) SaveLog(ctx context.Context, logEntry IncidentLog) error {
	log.Printf("[DEBIUG] Saving IncidentLog: %+v", logEntry)

	_, _, err := r.client.Collection(IncidentLogsCollection).Add(ctx, logEntry)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		return err
	}
	return nil
}

func (r *LogRepository) SaveMetric(ctx context.Context, metricEntry MetricLog) error {
	log.Printf("[DEBIUG] Saving MetricLog: %+v", metricEntry)

	_, _, err := r.client.Collection(MetricLogsCollection).Add(ctx, metricEntry)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		return err
	}
	return nil
}

func (r *LogRepository) GetIncidentsByService(ctx context.Context, serviceID uint) ([]IncidentLog, error) {
	var incidents []IncidentLog

	query := r.client.Collection(IncidentLogsCollection).
		Where("monitored_service_id", "==", int64(serviceID))

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		var incident IncidentLog
		if err := doc.DataTo(&incident); err != nil {
			log.Printf("Failed to map document %s to IncidentLog: %v", doc.Ref.ID, err)
			continue
		}
		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (r *LogRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
