output "frontend_ip_address" {
  value = google_compute_global_address.frontend_ip.address
}

output "backend_ip_address" {
  value = google_compute_global_address.backend_ip.address
}

output "redis_host" {
  value = google_redis_instance.redis.host
}
