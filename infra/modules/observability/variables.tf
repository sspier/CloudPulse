variable "env" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
}

variable "ecs_cluster_name" {
  description = "ECS cluster name for the API service"
  type        = string
}

variable "ecs_service_name" {
  description = "ECS service name for the API service"
  type        = string
}

variable "runner_function_name" {
  description = "Lambda function name for the runner"
  type        = string
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name for probe results"
  type        = string
}

variable "alarm_actions" {
  description = "List of ARNs to notify when alarms go into ALARM state (e.g., SNS topics)"
  type        = list(string)
  default     = []
}

variable "ok_actions" {
  description = "List of ARNs to notify when alarms return to OK state"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Base tags to apply to observability resources"
  type        = map(string)
  default     = {}
}
