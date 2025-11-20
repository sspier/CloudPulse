# prefix for the table name used by all environments
# final name: <prefix>-<env>
variable "table_name_prefix" {
  description = "base prefix for the dynamodb table name (e.g., cloudpulse-probe-results)"
  type        = string
}

# environment like dev or prod
variable "env" {
  description = "environment name (e.g., dev, prod)"
  type        = string
}

# common tags used across cloudpulse resources
variable "tags" {
  description = "tags to apply to the dynamodb table"
  type        = map(string)
  default     = {}
}
