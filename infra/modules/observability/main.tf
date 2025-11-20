locals {
  # shared prefix used for all observability resources in this environment
  name_prefix = "cloudpulse-${var.env}"
}

#############################
# ecs api cpu alarm
#############################

# alarms when the api ecs service averages >80% cpu for 3 minutes
# helps catch runaway traffic, bad deployments, or container loops
resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  alarm_name          = "${local.name_prefix}-ecs-api-high-cpu"
  alarm_description   = "cloudpulse api ecs service cpu > 80% for 3 minutes (${var.env})"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    ClusterName = var.ecs_cluster_name
    ServiceName = var.ecs_service_name
  }

  # missing data treated as healthy to avoid false positives after deploys
  treat_missing_data = "notBreaching"
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-ecs-api-high-cpu" }
  )
}

#############################
# runner lambda errors alarm
#############################

# triggers if the probe-runner lambda records any errors
# catches issues with network calls, bad target lists, or permissions
resource "aws_cloudwatch_metric_alarm" "runner_errors" {
  alarm_name        = "${local.name_prefix}-runner-errors"
  alarm_description = "cloudpulse runner lambda has errors (${var.env})"

  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 1

  dimensions = {
    FunctionName = var.runner_function_name
  }

  treat_missing_data = "notBreaching"
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-runner-errors" }
  )
}

#############################
# dynamodb throttling alarms
#############################

# alarms when reads on the probe-results table are throttled
# usually indicates the runner is reading too frequently or misconfigured
resource "aws_cloudwatch_metric_alarm" "ddb_read_throttles" {
  alarm_name        = "${local.name_prefix}-ddb-read-throttles"
  alarm_description = "dynamodb read throttles on probe results table (${var.env})"

  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "ReadThrottleEvents"
  namespace           = "AWS/DynamoDB"
  period              = 300
  statistic           = "Sum"
  threshold           = 1

  dimensions = {
    TableName = var.dynamodb_table_name
  }

  treat_missing_data = "notBreaching"
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-ddb-read-throttles" }
  )
}

# alarms when writes to the table are throttled
# helps detect issues where too many probe results are being written too fast
resource "aws_cloudwatch_metric_alarm" "ddb_write_throttles" {
  alarm_name        = "${local.name_prefix}-ddb-write-throttles"
  alarm_description = "dynamodb write throttles on probe results table (${var.env})"

  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "WriteThrottleEvents"
  namespace           = "AWS/DynamoDB"
  period              = 300
  statistic           = "Sum"
  threshold           = 1

  dimensions = {
    TableName = var.dynamodb_table_name
  }

  treat_missing_data = "notBreaching"
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-ddb-write-throttles" }
  )
}

#############################
# cloudwatch dashboard
#############################

# dashboard summarizing the health of the api, runner, and results table
# gives a quick at-a-glance view into system behavior
resource "aws_cloudwatch_dashboard" "this" {
  dashboard_name = "${local.name_prefix}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "ECS API CPU Utilization"
          view   = "timeSeries"
          region = "us-east-1"
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ClusterName", var.ecs_cluster_name, "ServiceName", var.ecs_service_name]
          ]
          stat = "Average"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "Runner Lambda Errors"
          view   = "timeSeries"
          region = "us-east-1"
          metrics = [
            ["AWS/Lambda", "Errors", "FunctionName", var.runner_function_name]
          ]
          stat = "Sum"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 24
        height = 6
        properties = {
          title  = "DynamoDB Throttled Events"
          view   = "timeSeries"
          region = "us-east-1"
          metrics = [
            ["AWS/DynamoDB", "ReadThrottleEvents", "TableName", var.dynamodb_table_name],
            ["...", "WriteThrottleEvents", "TableName", var.dynamodb_table_name]
          ]
          stat = "Sum"
        }
      }
    ]
  })
}
