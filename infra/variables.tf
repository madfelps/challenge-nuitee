variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "api_replicas" {
  description = "Number of API replicas"
  type        = number
  default     = 1
}

variable "database_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "nuitee"
}

variable "database_user" {
  description = "PostgreSQL database user"
  type        = string
  default     = "nuitee"
}

variable "database_password" {
  description = "PostgreSQL database password"
  type        = string
  sensitive   = true
}

variable "postgres_storage_size" {
  description = "PostgreSQL storage size"
  type        = string
  default     = "1Gi"
}

variable "storage_class_name" {
  description = "Kubernetes storage class name"
  type        = string
  default     = "standard"
}

variable "lite_api_url" {
  description = "LiteAPI base URL"
  type        = string
  default     = "https://api.liteapi.travel/v3.0"
}

variable "lite_api_key" {
  description = "LiteAPI key"
  type        = string
  sensitive   = true
}

variable "mailtrap_username" {
  description = "Mailtrap SMTP username"
  type        = string
  sensitive   = true
}

variable "mailtrap_password" {
  description = "Mailtrap SMTP password"
  type        = string
  sensitive   = true
}

