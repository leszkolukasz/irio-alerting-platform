output "registry_url" {
  value = google_artifact_registry_repository.docker_registry.registry_uri
}

output "bucket_name" {
  value = google_storage_bucket.static_assets.name
}
