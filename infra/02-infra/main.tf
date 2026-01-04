terraform {
  backend "gcs" {
    bucket = "alerting-platform-tf-state"
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

  enable_cdn = true

  # TODO: sprawdź ustawienia cdn
}

resource "google_compute_global_address" "frontend_ip" {
  name = "alerting-platform-frontend-ip"
}

resource "google_compute_global_address" "backend_ip" {
  name = "alerting-platform-backend-ip"
}


resource "google_compute_url_map" "url_map" {
  name            = "alerting-platform-url-map"
  default_service = google_compute_backend_bucket.static_assets_backend.id
}

resource "google_compute_target_http_proxy" "http_proxy" {
  name    = "alerting-platform-http-proxy"
  url_map = google_compute_url_map.url_map.id
}

resource "google_compute_global_forwarding_rule" "http_forwarding_rule" {
  name        = "alerting-platform-http-forwarding-rule"
  target      = google_compute_target_http_proxy.http_proxy.id
  port_range  = "80"
  ip_protocol = "TCP"

  ip_address = google_compute_global_address.frontend_ip.address
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
}

### Firestore
resource "google_firestore_database" "firestore_db" {
  project     = var.project_id
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

resource "google_project_service" "firestore_api" {
  project = var.project_id
  service = "firestore.googleapis.com"

  disable_on_destroy = false
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
}

resource "google_container_node_pool" "primary_nodes" {
  name    = "primary-node-pool"
  cluster = google_container_cluster.gke.name

  node_count = 1

  autoscaling {
    min_node_count = 1
    max_node_count = 1
  }

  node_config {
    machine_type = "e2-micro"
    spot         = true

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}
