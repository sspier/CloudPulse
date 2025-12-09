locals {
  # shared name prefix for all ecs/api resources in this module
  name_prefix = "cloudpulse-api"
}

# ecs cluster that will run the cloudpulse api service
resource "aws_ecs_cluster" "this" {
  name = "${local.name_prefix}-cluster"

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-cluster"
    }
  )
}

# cloudwatch log group for container logs from the ecs tasks
resource "aws_cloudwatch_log_group" "this" {
  name              = "/ecs/${local.name_prefix}"
  retention_in_days = 14

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-logs"
    }
  )
}

# security group for the alb
# allows http traffic from the internet in and all traffic out
resource "aws_security_group" "alb" {
  name        = "${local.name_prefix}-alb-sg"
  description = "alb security group for cloudpulse api"
  vpc_id      = var.vpc_id

  ingress {
    description = "http from the internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-alb-sg"
    }
  )
}

# security group for the ecs tasks
# only the alb is allowed to talk to the tasks on the container port
resource "aws_security_group" "ecs_tasks" {
  name        = "${local.name_prefix}-ecs-sg"
  description = "ecs tasks security group for cloudpulse api"
  vpc_id      = var.vpc_id

  ingress {
    description = "traffic from alb"
    from_port   = var.container_port
    to_port     = var.container_port
    protocol    = "tcp"
    security_groups = [
      aws_security_group.alb.id
    ]
  }

  egress {
    description = "all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-ecs-sg"
    }
  )
}

# application load balancer in the public subnets
# fronts the cloudpulse api service running on ecs
resource "aws_lb" "this" {
  name               = "${local.name_prefix}-alb"
  load_balancer_type = "application"
  internal           = false
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-alb"
    }
  )
}

# target group for the ecs tasks
# uses ip mode so each task’s eni can be registered directly
resource "aws_lb_target_group" "this" {
  name        = "${local.name_prefix}-tg"
  port        = var.container_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = var.vpc_id

  health_check {
    path                = "/health"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 15
    matcher             = "200"
  }

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-tg"
    }
  )
}

# http listener on port 80 that forwards all traffic to the ecs target group
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this.arn
  }
}

# iam trust policy for ecs tasks to assume the execution and task roles
data "aws_iam_policy_document" "ecs_task_execution_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# execution role used by ecs to pull images, write logs, etc
resource "aws_iam_role" "ecs_task_execution" {
  name               = "${local.name_prefix}-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_assume_role.json

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-execution-role"
    }
  )
}

# attach the standard ecs task execution managed policy
resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# task role for the cloudpulse api containers
# this is where app-level permissions will go (dynamodb, xray, etc)
resource "aws_iam_role" "ecs_task" {
  name               = "${local.name_prefix}-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_assume_role.json

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-task-role"
    }
  )
}

# grant the ecs task permission to read/write to the dynamodb tables
resource "aws_iam_role_policy" "dynamodb_access" {
  name = "${local.name_prefix}-dynamodb-policy"
  role = aws_iam_role.ecs_task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchWriteItem",
          "dynamodb:BatchGetItem"
        ]
        Resource = [
          "arn:aws:dynamodb:*:*:table/${var.table_name_targets}",
          "arn:aws:dynamodb:*:*:table/${var.table_name_results}",
          "arn:aws:dynamodb:*:*:table/${var.table_name_results}/index/*"
        ]
      }
    ]
  })
}

# ecs task definition for the cloudpulse api fargate task
# wires up the container image, port, env vars, logs, and cpu/memory
resource "aws_ecs_task_definition" "this" {
  family                   = "${local.name_prefix}-task"
  cpu                      = "256"
  memory                   = "512"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name      = "cloudpulse-api"
      image     = var.container_image
      essential = true
      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "PORT"
          value = tostring(var.container_port)
        },
        {
          name  = "TABLE_NAME_TARGETS"
          value = var.table_name_targets
        },
        {
          name  = "TABLE_NAME_RESULTS"
          value = var.table_name_results
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.this.name
          awslogs-region        = var.region
          awslogs-stream-prefix = "ecs"
        }
      }
    }
  ])

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-task-def"
    }
  )
}

# ecs service that runs the cloudpulse api tasks behind the alb
resource "aws_ecs_service" "this" {
  name            = "${local.name_prefix}-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.this.arn
    container_name   = "cloudpulse-api"
    container_port   = var.container_port
  }

  # ignore_changes on task_definition lets us do rolling deploys
  # by creating new task defs without terraform constantly diffing them
  lifecycle {
    ignore_changes = [task_definition]
  }

  depends_on = [
    aws_lb_listener.http,
  ]

  tags = merge(
    var.tags,
    {
      Name = "${local.name_prefix}-service"
    }
  )
}
