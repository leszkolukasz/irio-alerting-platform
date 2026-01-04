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

resource "google_pubsub_topic" "service_created" {
  name = "service-created"
}

resource "google_pubsub_topic" "service_removed" {
  name = "service-removed"
}

resource "google_pubsub_topic" "service_modified" {
  name = "service-modified"
}

resource "google_pubsub_topic" "notify_oncaller" {
  name = "notify-oncaller"
}

resource "google_pubsub_topic" "oncaller_acknowledged" {
  name = "oncaller-acknowledged"
}

# Subscriptions
resource "google_pubsub_subscription" "logger_incident_start" {
  name  = "logger-incident-start"
  topic = google_pubsub_topic.incident_start.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_incident_resolved" {
  name  = "logger-incident-resolved"
  topic = google_pubsub_topic.incident_resolved.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_incident_timeout" {
  name  = "logger-incident-timeout"
  topic = google_pubsub_topic.incident_timeout.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_incident_unresolved" {
  name  = "logger-incident-unresolved"
  topic = google_pubsub_topic.incident_unresolved.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_service_up" {
  name  = "logger-service-up"
  topic = google_pubsub_topic.service_up.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_service_down" {
  name  = "logger-service-down"
  topic = google_pubsub_topic.service_down.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "logger_notify_oncaller" {
  name  = "logger-notify-oncaller"
  topic = google_pubsub_topic.notify_oncaller.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_service_up" {
  name  = "incident-manager-service-up"
  topic = google_pubsub_topic.service_up.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_service_down" {
  name  = "incident-manager-service-down"
  topic = google_pubsub_topic.service_down.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_service_created" {
  name  = "incident-manager-service-created"
  topic = google_pubsub_topic.service_created.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_service_removed" {
  name  = "incident-manager-service-removed"
  topic = google_pubsub_topic.service_removed.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_service_modified" {
  name  = "incident-manager-service-modified"
  topic = google_pubsub_topic.service_modified.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "incident_manager_oncaller_acknowledged" {
  name  = "incident-manager-oncaller-acknowledged"
  topic = google_pubsub_topic.oncaller_acknowledged.name

  enable_message_ordering = true
}

resource "google_pubsub_subscription" "notifier_notify_oncaller" {
  name  = "notifier-notify-oncaller"
  topic = google_pubsub_topic.notify_oncaller.name

  enable_message_ordering = true
}
