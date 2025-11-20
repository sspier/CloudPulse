terraform {
  # require a reasonably recent terraform version
  required_version = ">= 1.11.0"

  # bring in the aws provider at a stable version
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # store terraform state in s3 so multiple environments and users can share state safely
  # locking prevents concurrent applies from corrupting state
  backend "s3" {
    bucket       = "cloudpulse-tf-state"
    key          = "envs/dev/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true # encrypt state at rest
    use_lockfile = true # enable state locking with s3 + dynamodb
  }
}

# default aws provider for this environment
provider "aws" {
  region = "us-east-1"
}

# identity and region info used for building ecr image urls and tagging
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ecr repo used to store the cloudpulse api container image
resource "aws_ecr_repository" "api" {
  name = "cloudpulse-api"

  image_scanning_configuration {
    scan_on_push = true # enable vulnerability scans on new images
  }

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

# convenience local for the full ecr image url for the api
locals {
  api_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_repository.api.name}:dev"
}

# vpc module creates network foundation: vpc, subnets, route tables, etc
module "vpc" {
  source = "../../modules/vpc"

  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.0.0/24", "10.0.1.0/24"]
  private_subnet_cidrs = ["10.0.2.0/24", "10.0.3.0/24"]

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

# ecs service running the cloudpulse api
# deploys an ecs cluster, task definition, load balancer, etc
module "ecs_api" {
  source = "../../modules/ecs_api"

  vpc_id             = module.vpc.vpc_id
  public_subnet_ids  = module.vpc.public_subnet_ids
  private_subnet_ids = module.vpc.private_subnet_ids

  container_image = local.api_image
  container_port  = 8080
  desired_count   = 1
  region          = data.aws_region.current.name

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

# dynamodb table for storing probe results
# this gives the api somewhere durable to write uptime checks
module "results_table" {
  source = "../../modules/dynamodb_results"

  table_name_prefix = "cloudpulse-probe-results"
  env               = "dev"

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

# lambda function responsible for background probing in the cloud deployment
# eventually replaces the local go scheduler used in dev
module "runner" {
  source = "../../modules/runner"

  env                 = "dev"
  schedule_expression = "rate(1 minute)" # how often the runner should probe targets

  # placeholder image until runner container exists
  runner_image = "123456789012.dkr.ecr.us-east-1.amazonaws.com/cloudpulse-runner:dev"

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

# observability stack for cloudpulse
# creates alarms, dashboards, metrics, etc. wired to ecs and lambda
module "observability" {
  source = "../../modules/observability"

  env                  = "dev"
  ecs_cluster_name     = module.ecs_api.ecs_cluster_name
  ecs_service_name     = module.ecs_api.ecs_service_name
  runner_function_name = module.runner.lambda_function_name
  dynamodb_table_name  = module.results_table.table_name

  alarm_actions = []
  ok_actions    = []

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}
