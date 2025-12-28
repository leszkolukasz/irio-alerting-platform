package db

import (
	"time"

	"cloud.google.com/go/firestore"
)

type LogRepository struct {
	client *firestore.Client
}

type IncidentLog struct {
	IncidentID   string                 `firestore:"incident_id"`
	OnCallerData map[string]interface{} `firestore:"oncaller_data, omitempty"`
	Timestamp    time.Time              `firestore:"timestamp"`
	Type         string                 `firestore:"type"`
}

type MetricLog struct {
	ServiceID string    `firestore:"monitored_service_id"`
	Timestamp time.Time `firestore:"timestamp"`
	Type      string    `firestore:"type"`
}
