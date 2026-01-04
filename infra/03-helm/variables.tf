variable "project_id" {
  type = string
}

variable "secret" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "europe-west1"
}

variable "firestore_db" {
  type    = string
  default = "logger-db"
}

variable "app_version" {
  type    = string
  default = "v1.0.0"
}

variable "build_time" {
  type    = string
  default = "2026-01-01T00:00:00Z"
}

variable "smtp_host" {
  type = string
}

variable "smtp_port" {
  type    = number
  default = 2525
}

variable "smtp_user" {
  type = string
}

variable "smtp_pass" {
  type      = string
  sensitive = true
}

variable "email_from" {
  type    = string
  default = "alerting-service@test-y7zpl983d6o45vx6.mlsender.net"
}
