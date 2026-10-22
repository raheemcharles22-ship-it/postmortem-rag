variable "app_image" {
  description = "Container image for the app"
  type        = string
  default     = "postmortem-rag:local"
}

variable "openai_api_key" {
  description = "OpenAI API key"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Postgres password"
  type        = string
  sensitive   = true
}