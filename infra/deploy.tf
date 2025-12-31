data "google_client_config" "default" {}

provider "helm" {
  kubernetes = {
    host                   = "https://${google_container_cluster.gke.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(google_container_cluster.gke.master_auth[0].cluster_ca_certificate)
  }
}

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
      value = google_artifact_registry_repository.docker_registry.registry_uri
    },
    {
      name  = "env.FIRESTORE_DB"
      value = var.firestore_db
    }
  ]


  set_sensitive = [
    {
      name  = "secrets.POSTGRES_PASSWORD"
      value = google_sql_user.db_user.password
    },
    {
      name  = "secrets.REDIS_PASSWORD"
      value = google_redis_instance.redis.auth_string
    }
  ]
}
