package firestore

import (
	"time"

	"cloud.google.com/go/firestore"
)

type LogRepository struct {
	client *firestore.Client
}

type IncidentLog struct {
	IncidentID string    `firestore:"incident_id"`
	ServiceID  int64     `firestore:"monitored_service_id"`
	Oncaller   string    `firestore:"oncaller,omitempty"`
	Timestamp  time.Time `firestore:"timestamp"`
	Type       string    `firestore:"type"`
}

type MetricLog struct {
	ServiceID string    `firestore:"monitored_service_id"`
	Timestamp time.Time `firestore:"timestamp"`
	Type      string    `firestore:"type"`
}
