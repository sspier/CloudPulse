output "dashboard_name" {
  description = "Name of the CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.this.dashboard_name
}

output "ecs_cpu_alarm_name" {
  description = "Name of the ECS high CPU alarm"
  value       = aws_cloudwatch_metric_alarm.ecs_cpu_high.alarm_name
}
