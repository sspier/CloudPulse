variable "env" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
}

variable "runner_image" {
  description = "Container image URI for the runner Lambda (package_type=Image)"
  type        = string
}

variable "schedule_expression" {
  description = "EventBridge schedule expression for runner execution"
  type        = string
  default     = "rate(1 minute)"
}

variable "table_name_targets" {
  description = "Name of the targets DynamoDB table"
  type        = string
}

variable "table_name_results" {
  description = "Name of the results DynamoDB table"
  type        = string
}

variable "tags" {
  description = "Base tags to apply to runner resources"
  type        = map(string)
  default     = {}
}
