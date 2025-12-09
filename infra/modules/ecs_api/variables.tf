variable "vpc_id" {
  description = "ID of the VPC where ECS and ALB will run"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the ALB"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "container_image" {
  description = "Container image URI for the CloudPulse API"
  type        = string
}

variable "container_port" {
  description = "Port the API container listens on"
  type        = number
  default     = 8080
}

variable "desired_count" {
  description = "Desired number of ECS tasks"
  type        = number
  default     = 2
}

variable "region" {
  description = "AWS region (used for logs)"
  type        = string
  default     = "us-east-1"
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
  description = "Base tags to apply to ECS and ALB resources"
  type        = map(string)
  default     = {}
}
