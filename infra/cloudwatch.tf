# CloudWatch resources for observability
# These resources are only created in AWS (not local development)

# Log group for API logs
resource "aws_cloudwatch_log_group" "api" {
  count = var.is_local ? 0 : 1

  name              = "/app/${var.environment}/api"
  retention_in_days = 14

  tags = {
    Environment = var.environment
    Service     = "api"
  }
}

# SNS topic for alerts
resource "aws_sns_topic" "alerts" {
  count = var.is_local ? 0 : 1

  name = "api-alerts-${var.environment}"

  tags = {
    Environment = var.environment
  }
}

# Email subscription for alerts (optional - requires confirmation)
resource "aws_sns_topic_subscription" "alert_email" {
  count = var.is_local || var.alert_email == "" ? 0 : 1

  topic_arn = aws_sns_topic.alerts[0].arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# Alarm: High error rate
resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  count = var.is_local ? 0 : 1

  alarm_name          = "api-high-error-rate-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ErrorCount"
  namespace           = var.cloudwatch_namespace
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "API error count exceeded threshold"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alerts[0].arn]
  ok_actions    = [aws_sns_topic.alerts[0].arn]

  tags = {
    Environment = var.environment
  }
}

# Alarm: High latency (p95)
resource "aws_cloudwatch_metric_alarm" "high_latency" {
  count = var.is_local ? 0 : 1

  alarm_name          = "api-high-latency-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "RequestLatency"
  namespace           = var.cloudwatch_namespace
  period              = 300
  extended_statistic  = "p95"
  threshold           = 2000
  alarm_description   = "API p95 latency exceeded 2 seconds"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alerts[0].arn]
  ok_actions    = [aws_sns_topic.alerts[0].arn]

  tags = {
    Environment = var.environment
  }
}
