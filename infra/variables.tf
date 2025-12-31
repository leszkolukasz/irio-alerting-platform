variable "db_user" {
  type = string
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "project_id" {
  type    = string
  default = "fill_this"
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
