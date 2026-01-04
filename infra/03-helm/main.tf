terraform {
  backend "gcs" {
    bucket = "alerting-platform-tf-state"
    prefix = "terraform/helm"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 7.0.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 3.0.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

## Data

data "google_client_config" "default" {}

data "google_container_cluster" "gke" {
  name     = "alerting-platform-gke"
  location = var.region
}

data "google_redis_instance" "redis" {
  name   = "alerting-platform-redis"
  region = var.region
}

data "google_sql_database_instance" "db_instance" {
  name = "alerting-platform-db"
}

data "google_artifact_registry_repository" "docker_registry" {
  location      = var.region
  repository_id = "docker-registry"
}

data "google_compute_global_address" "backend_ip" {
  name = "alerting-platform-backend-ip"
}

provider "helm" {
  kubernetes = {
    host                   = "https://${google_container_cluster.gke.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(google_container_cluster.gke.master_auth[0].cluster_ca_certificate)
  }
}

## Deploy

resource "helm_release" "alerting-platform" {
  name             = "alerting-platform"
  repository       = "./charts"
  chart            = "alerting-platform"
  namespace        = "default"
  create_namespace = true

  timeout = 600
  atomic  = true


  set = [
    {
      name  = "env.VERSION"
      value = var.app_version
    },
    {
      name  = "env.BUILD_TIME"
      value = var.build_time
    },
    {
      name  = "env.POSTGRES_USER"
      value = google_sql_user.db_user.name
    },
    {
      name  = "env.POSTGRES_DB"
      value = google_sql_database.api_db.name
    },
    {
      name  = "env.REDIS_HOST"
      value = google_redis_instance.redis.host
    },
    {
      name  = "env.REDIS_PORT"
      value = google_redis_instance.redis.port
    },
    {
      name  = "env.PROJECT_ID"
      value = var.project_id
    },
    {
      name  = "gcloud.projectID"
      value = var.project_id
    },
    {
      name  = "gcloud.region"
      value = var.region
    },
    {
      name  = "gcloud.dbInstance"
      value = google_sql_database_instance.db_instance.name
    },
    {
      name  = "gcloud.registryURL"
      value = data.google_artifact_registry_repository.docker_registry.registry_uri
    },
    {
      name  = "env.FIRESTORE_DB"
      value = var.firestore_db
    },
    {
      name  = "env.SMTP_HOST"
      value = var.smtp_host
    },
    {
      name  = "env.SMTP_PORT"
      value = var.smtp_port
    },
    {
      name  = "env.SMTP_USER"
      value = var.smtp_user
    },
    {
      name  = "env.EMAIL_FROM"
      value = var.email_from
    },
    {
      name  = "gcloud.publicAPIURL"
      value = "http://${google_compute_global_address.backend_ip.address}"
    }
  ]


  set_sensitive = [
    {
      name  = "secrets.SECRET",
      value = var.secret
    },
    {
      name  = "secrets.POSTGRES_PASSWORD"
      value = google_sql_user.db_user.password
    },
    {
      name  = "secrets.REDIS_PASSWORD"
      value = google_redis_instance.redis.auth_string
    },
    {
      name  = "secrets.SMTP_PASS"
      value = var.smtp_pass
    }
  ]
}
