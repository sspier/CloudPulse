locals {
  # shared prefix used for all runner resources for this environment
  name_prefix = "cloudpulse-runner-${var.env}"
}

# iam trust policy so the lambda service can assume the runner role
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# base lambda execution role
# allows writing logs and running in the default lambda environment
resource "aws_iam_role" "lambda_role" {
  name               = "${local.name_prefix}-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-role" }
  )
}

# attach aws-managed policy for basic lambda execution
# this gives permissions for cloudwatch logs, which all lambdas need
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# runner lambda that executes the probe logic on a schedule
# container image is built separately and pushed to ecr
resource "aws_lambda_function" "runner" {
  function_name = local.name_prefix
  role          = aws_iam_role.lambda_role.arn

  package_type = "Image" # use container image instead of zip
  image_uri    = var.runner_image

  timeout     = 30  # enough time to probe multiple targets
  memory_size = 256 # modest memory footprint

  environment {
    variables = {
      ENV = var.env # pass environment context to the container
    }
  }

  tags = merge(
    var.tags,
    { Name = local.name_prefix }
  )
}

# eventbridge rule that determines how often the runner executes
# supports flexible expressions like rate(1 minute) or cron(...)
resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "${local.name_prefix}-schedule"
  schedule_expression = var.schedule_expression

  tags = merge(
    var.tags,
    { Name = "${local.name_prefix}-schedule" }
  )
}

# eventbridge target that wires the schedule to the lambda function
resource "aws_cloudwatch_event_target" "lambda_target" {
  rule      = aws_cloudwatch_event_rule.schedule.name
  target_id = "${local.name_prefix}-target"
  arn       = aws_lambda_function.runner.arn
}

# allow eventbridge to invoke the lambda
# lambda permissions must explicitly trust the calling service
resource "aws_lambda_permission" "allow_events" {
  statement_id  = "AllowExecutionFromEventBridge-${var.env}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.runner.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}
