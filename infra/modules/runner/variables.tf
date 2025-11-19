variable "env" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
}

variable "runner_image" {
  description = "Container image URI for the runner Lambda (package_type=Image)"
  type        = string
}

variable "schedule_expression" {
  description = "EventBridge schedule expression, e.g. rate(1 minute)"
  type        = string
  default     = "rate(1 minute)"
}

variable "tags" {
  description = "Base tags to apply to runner resources"
  type        = map(string)
  default     = {}
}
