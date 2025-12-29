# Topics
resource "google_pubsub_topic" "incident_start" {
  name = "incident-start"
}

resource "google_pubsub_topic" "incident_resolved" {
  name = "incident-resolved"
}

resource "google_pubsub_topic" "incident_timeout" {
  name = "incident-acknowledge-timeout"
}

resource "google_pubsub_topic" "incident_unresolved" {
  name = "incident-unresolved"
}

resource "google_pubsub_topic" "service_up" {
  name = "service-up"
}

resource "google_pubsub_topic" "service_down" {
  name = "service-down"
}

resource "google_pubsub_topic" "execute_health_check" {
  name = "execute-health-check"
}

resource "google_pubsub_topic" "alert_created" {
  name = "alert-created"
}

resource "google_pubsub_topic" "alert_removed" {
  name = "alert-removed"
}

resource "google_pubsub_topic" "alert_modified" {
  name = "alert-modified"
}

resource "google_pubsub_topic" "notify_oncaller" {
  name = "notify-oncaller"
}

# Subscriptions
resource "google_pubsub_subscription" "logger_incident_start" {
  name  = "logger-incident-start"
  topic = google_pubsub_topic.incident_start.name
}

resource "google_pubsub_subscription" "logger_incident_resolved" {
  name  = "logger-incident-resolved"
  topic = google_pubsub_topic.incident_resolved.name
}

resource "google_pubsub_subscription" "logger_incident_timeout" {
  name  = "logger-incident-timeout"
  topic = google_pubsub_topic.incident_timeout.name
}

resource "google_pubsub_subscription" "logger_incident_unresolved" {
  name  = "logger-incident-unresolved"
  topic = google_pubsub_topic.incident_unresolved.name
}

resource "google_pubsub_subscription" "logger_service_up" {
  name  = "logger-service-up"
  topic = google_pubsub_topic.service_up.name
}

resource "google_pubsub_subscription" "logger_service_down" {
  name  = "logger-service-down"
  topic = google_pubsub_topic.service_down.name
}