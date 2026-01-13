terraform {
  backend "gcs" {
    prefix = "terraform/infra"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 7.0.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

## Load Balancer

data "google_storage_bucket" "static_assets" {
  name = "alerting-platform-static-assets-${var.project_id}"
}

resource "google_compute_backend_bucket" "static_assets_backend" {
  name        = "static-assets-backend"
  bucket_name = data.google_storage_bucket.static_assets.name

  enable_cdn = false
}

data "google_compute_global_address" "frontend_ip" {
  name = "alerting-platform-frontend-ip"
}

data "google_compute_global_address" "backend_ip" {
  name = "alerting-platform-backend-ip"
}


resource "google_compute_url_map" "url_map" {
  name            = "alerting-platform-url-map"
  default_service = google_compute_backend_bucket.static_assets_backend.id
}

resource "google_compute_managed_ssl_certificate" "frontend_cert" {
  name = "frontend-cert"

  managed {
    domains = ["alerting-platform.leszko.dev"]
  }
}

resource "google_compute_target_http_proxy" "http_proxy" {
  name    = "alerting-platform-http-proxy"
  url_map = google_compute_url_map.url_map.id
}

resource "google_compute_target_https_proxy" "https_proxy" {
  name             = "alerting-platform-https-proxy"
  url_map          = google_compute_url_map.url_map.id
  ssl_certificates = [google_compute_managed_ssl_certificate.frontend_cert.id]
}

resource "google_compute_global_forwarding_rule" "http_forwarding_rule" {
  name        = "alerting-platform-http-forwarding-rule"
  target      = google_compute_target_http_proxy.http_proxy.id
  port_range  = "80"
  ip_protocol = "TCP"

  ip_address = data.google_compute_global_address.frontend_ip.address
}

resource "google_compute_global_forwarding_rule" "https_forwarding_rule" {
  name        = "alerting-platform-https-forwarding-rule"
  target      = google_compute_target_https_proxy.https_proxy.id
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = data.google_compute_global_address.frontend_ip.address
}

## Database

resource "google_sql_database_instance" "db_instance" {
  name             = "alerting-platform-db"
  database_version = "POSTGRES_17"

  deletion_protection = false

  settings {
    tier    = "db-f1-micro"
    edition = "ENTERPRISE"
  }
}

resource "google_sql_user" "db_user" {
  instance = google_sql_database_instance.db_instance.name
  name     = var.db_user
  password = var.db_password
}

resource "google_sql_database" "api_db" {
  name     = "alerting_platform_api"
  instance = google_sql_database_instance.db_instance.name

  # Ir is not destroyed withouth this dependency
  depends_on = [google_project_service.sql_admin_api]
}

### Firestore
resource "google_firestore_database" "firestore_db" {
  name        = var.firestore_db
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  delete_protection_state = "DELETE_PROTECTION_DISABLED"
  deletion_policy         = "DELETE"
}

# Without this you cannot connect to Cloud SQL
resource "google_project_service" "sql_admin_api" {
  service = "sqladmin.googleapis.com"

  disable_on_destroy = false
}

resource "google_firestore_index" "metric_logs_composite" {
  database   = google_firestore_database.firestore_db.name
  collection = "metric_logs"

  fields {
    field_path = "monitored_service_id"
    order      = "ASCENDING"
  }

  fields {
    field_path = "timestamp"
    order      = "ASCENDING"
  }
}

resource "google_redis_instance" "redis" {
  name = "alerting-platform-redis"
  tier = "BASIC"

  auth_enabled        = true
  memory_size_gb      = 1
  deletion_protection = false
}

## GKE

resource "google_container_cluster" "gke" {
  name = "alerting-platform-gke"

  initial_node_count       = 1
  remove_default_node_pool = true

  deletion_protection = false

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
  }
  
  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
}

resource "google_container_node_pool" "primary_nodes" {
  name    = "primary-node-pool"
  cluster = google_container_cluster.gke.name

  node_count = 1

  # In total there is num_zones * node_count nodes created
  autoscaling {
    min_node_count = 1
    max_node_count = 3
  }

  node_config {
    machine_type = "e2-small"
    spot         = true

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}

## MONITORING & ALERTING
resource "google_monitoring_notification_channel" "email_channel" {
  display_name = "Email Notification Channel"
  type         = "email"
  
  labels = {
    email_address = var.monitoring_mail
  }
}

resource "google_monitoring_alert_policy" "container_restart_alert" {
  display_name = "GKE Container Restarting"
  combiner     = "OR"
  
  conditions {
    display_name = "Container Restarting > 0"
    
    condition_threshold {
      filter = "resource.type = \"k8s_container\" AND metric.type = \"kubernetes.io/container/restart_count\""
      
      duration   = "0s"
      
      comparison = "COMPARISON_GT"
      
      aggregations {
        alignment_period   = "600s"
        per_series_aligner = "ALIGN_DELTA"
      }
      
      threshold_value = 2
    }
  }

  notification_channels = [google_monitoring_notification_channel.email_channel.name]
  
  alert_strategy {
    auto_close = "1800s"
  }
}

resource "google_monitoring_alert_policy" "high_cpu_alert" {
  display_name = "High Node CPU Usage"
  combiner     = "OR"

  conditions {
    display_name = "Node CPU > 80%"
    
    condition_threshold {
      filter = "resource.type = \"k8s_node\" AND metric.type = \"kubernetes.io/node/cpu/allocatable_utilization\""
      
      duration   = "120s"
      comparison = "COMPARISON_GT"
      
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
      
      threshold_value = 0.8
    }
  }

  notification_channels = [google_monitoring_notification_channel.email_channel.name]

    
  alert_strategy {
    auto_close = "1800s"
  }
}

resource "google_logging_metric" "api_500_errors" {
  name        = "api_500_errors_metric"
  description = "500 Api Errors Count"

  filter = <<-EOT
    resource.type="k8s_container"
    AND resource.labels.container_name="alerting-platform-api"
    AND logName =~ "(stdout|stderr)"
    AND textPayload =~ "[|] 500 [|]"
  EOT

  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_monitoring_alert_policy" "api_500_alert" {
  display_name = "API Critical: HTTP 500 Errors"
  combiner     = "OR"
  enabled      = true

  conditions {
    display_name = "API 500 Errors > 1/min"
    
    condition_threshold {
      filter = "resource.type = \"k8s_container\" AND metric.type = \"logging.googleapis.com/user/${google_logging_metric.api_500_errors.name}\""
      
      duration   = "0s"
      comparison = "COMPARISON_GT"
      
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_SUM"
        
        cross_series_reducer = "REDUCE_SUM"
        
        group_by_fields = [
          "resource.label.cluster_name",
          "resource.label.namespace_name",
          "resource.label.container_name"
        ]
      }

      threshold_value = 5
    }
  }

  notification_channels = [google_monitoring_notification_channel.email_channel.name]
  
  alert_strategy {
    auto_close = "1800s"
  }
}

resource "google_monitoring_dashboard" "alerting_platform_dashboard" {
  dashboard_json = <<EOF
{
  "displayName": "Alerting Platform - Health & Status",
  "gridLayout": {
    "columns": "2",
    "widgets": [
      {
        "title": "API 500 Errors (Log Metric)",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "resource.type=\"k8s_container\" metric.type=\"logging.googleapis.com/user/${google_logging_metric.api_500_errors.name}\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s"
                }
              }
            },
            "plotType": "STACKED_BAR",
            "minAlignmentPeriod": "60s"
          }],
          "yAxis": {
            "label": "Count",
            "scale": "LINEAR"
          }
        }
      },
      {
        "title": "Container Restarts (Namespace: default)",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "resource.type=\"k8s_container\" metric.type=\"kubernetes.io/container/restart_count\" resource.label.namespace_name=\"default\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_DELTA",
                  "alignmentPeriod": "300s"
                }
              }
            },
            "plotType": "LINE"
          }],
          "yAxis": {
            "label": "Restarts/s",
            "scale": "LINEAR"
          }
        }
      },
      {
        "title": "Node CPU Utilization",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "resource.type=\"k8s_node\" metric.type=\"kubernetes.io/node/cpu/allocatable_utilization\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_MEAN",
                  "alignmentPeriod": "300s"
                }
              }
            },
            "plotType": "LINE"
          }],
          "yAxis": {
            "label": "Utilization (0-1)",
            "scale": "LINEAR"
          }
        }
      }
    ]
  }
}
EOF
}