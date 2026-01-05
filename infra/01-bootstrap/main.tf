terraform {
  backend "gcs" {
    prefix = "terraform/bootstrap"
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

resource "google_artifact_registry_repository" "docker_registry" {
  repository_id = "docker-registry"
  format        = "DOCKER"
}

## Bucket

resource "google_storage_bucket" "static_assets" {
  name     = "alerting-platform-static-assets-${var.project_id}"
  location = "EU"

  force_destroy               = true
  uniform_bucket_level_access = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "index.html"
  }
}

resource "google_storage_bucket_iam_member" "public_bucket_access" {
  bucket = google_storage_bucket.static_assets.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_compute_global_address" "frontend_ip" {
  name = "alerting-platform-frontend-ip"
}

resource "google_compute_global_address" "backend_ip" {
  name = "alerting-platform-backend-ip"
}

# Enable APIs

resource "google_project_service" "cloudresourcemanager" {
  project = var.project_id
  service = "cloudresourcemanager.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "serviceusage" {
  project            = var.project_id
  service            = "serviceusage.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "firestore" {
  project = var.project_id
  service = "firestore.googleapis.com"

  disable_on_destroy = false

  depends_on = [
    google_project_service.cloudresourcemanager,
    google_project_service.serviceusage
  ]
}
