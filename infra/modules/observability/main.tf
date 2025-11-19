locals {
  name_prefix = "cloudpulse-${var.env}"
}

#############################
# ECS API CPU alarm
#############################

resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  alarm_name          = "${local.name_prefix}-ecs-api-high-cpu"
  alarm_description   = "CloudPulse API ECS service CPU > 80% for 3 minutes (${var.env})"
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

  treat_missing_data = "notBreaching"
  alarm_actions      = var.alarm_actions
  ok_actions         = var.ok_actions

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-ecs-api-high-cpu"
    }
  )
}

#############################
# Runner Lambda errors alarm
#############################

resource "aws_cloudwatch_metric_alarm" "runner_errors" {
  alarm_name        = "${local.name_prefix}-runner-errors"
  alarm_description = "CloudPulse runner Lambda has errors (${var.env})"

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
    {
      Name = "${local.name_prefix}-runner-errors"
    }
  )
}

#############################
# DynamoDB throttling alarms
#############################

resource "aws_cloudwatch_metric_alarm" "ddb_read_throttles" {
  alarm_name        = "${local.name_prefix}-ddb-read-throttles"
  alarm_description = "DynamoDB read throttles on probe results table (${var.env})"

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
    {
      Name = "${local.name_prefix}-ddb-read-throttles"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "ddb_write_throttles" {
  alarm_name        = "${local.name_prefix}-ddb-write-throttles"
  alarm_description = "DynamoDB write throttles on probe results table (${var.env})"

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
    {
      Name = "${local.name_prefix}-ddb-write-throttles"
    }
  )
}

#############################
# CloudWatch dashboard
#############################

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
