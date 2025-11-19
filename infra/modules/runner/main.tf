locals {
  name_prefix = "cloudpulse-runner-${var.env}"
}

// IAM assume-role policy for Lambda
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

// Basic execution role (logs, etc.)
resource "aws_iam_role" "lambda_role" {
  name               = "${local.name_prefix}-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-role"
    }
  )
}

// Attach AWS-managed policy for basic Lambda execution (CloudWatch logs)
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

// Runner Lambda function (stub for now, using container image)
resource "aws_lambda_function" "runner" {
  function_name = local.name_prefix
  role          = aws_iam_role.lambda_role.arn

  package_type = "Image"
  image_uri    = var.runner_image

  timeout     = 30
  memory_size = 256

  environment {
    variables = {
      ENV = var.env
    }
  }

  tags = merge(
    var.tags,
    {
      Name = local.name_prefix
    }
  )
}

// EventBridge rule that triggers the runner on a schedule
resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "${local.name_prefix}-schedule"
  schedule_expression = var.schedule_expression

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-schedule"
    }
  )
}

// EventBridge target: invoke the Lambda
resource "aws_cloudwatch_event_target" "lambda_target" {
  rule      = aws_cloudwatch_event_rule.schedule.name
  target_id = "${local.name_prefix}-target"
  arn       = aws_lambda_function.runner.arn
}

// Allow EventBridge to invoke the Lambda
resource "aws_lambda_permission" "allow_events" {
  statement_id  = "AllowExecutionFromEventBridge-${var.env}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.runner.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}
