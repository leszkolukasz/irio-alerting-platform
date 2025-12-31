output "frontend_ip_address" {
  value = google_compute_global_address.frontend_ip.address
}
output "backend_ip_address" {
  value = google_compute_global_address.backend_ip.address
}

output "bucket_name" {
  value = google_storage_bucket.static_assets.name
}

output "registry_url" {
  value = google_artifact_registry_repository.docker_registry.registry_uri
}

output "redis_host" {
  value = google_redis_instance.redis.host
}
