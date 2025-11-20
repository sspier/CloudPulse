terraform {
  # require a reasonably recent terraform version
  required_version = ">= 1.11.0"

  # aws provider pinned to a stable 5.x release
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # store production state in s3 with locking enabled
  # prod uses its own tfstate file to keep environments isolated
  backend "s3" {
    bucket       = "cloudpulse-tf-state"
    key          = "envs/prod/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

# default aws provider config for this environment
provider "aws" {
  region = "us-east-1"
}

# identity + region info used to build ecr urls and tags
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# full ecr path for the production api image
locals {
  api_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/cloudpulse-api:prod"
}

# ecs service running the cloudpulse api
# identical to dev but pulls the prod-tagged container image
module "ecs_api" {
  source = "../../modules/ecs_api"

  vpc_id             = module.vpc.vpc_id
  public_subnet_ids  = module.vpc.public_subnet_ids
  private_subnet_ids = module.vpc.private_subnet_ids

  container_image = local.api_image # prod image tag
  container_port  = 8080
  desired_count   = 1 # can scale higher later
  region          = data.aws_region.current.name

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}

# vpc and networking infra for prod
# same layout as dev for now; can diverge later if needed
module "vpc" {
  source = "../../modules/vpc"

  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.0.0/24", "10.0.1.0/24"]
  private_subnet_cidrs = ["10.0.2.0/24", "10.0.3.0/24"]

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}

# production dynamodb table for probe results
# same schema as dev but isolated for safety and cleaner deployments
module "results_table" {
  source = "../../modules/dynamodb_results"

  table_name_prefix = "cloudpulse-probe-results"
  env               = "prod"

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}

# production runner lambda for probing targets on a schedule
# prod probes less frequently than dev for cost and stability
module "runner" {
  source = "../../modules/runner"

  env                 = "prod"
  schedule_expression = "rate(5 minutes)" # slower cycle for prod

  runner_image = "123456789012.dkr.ecr.us-east-1.amazonaws.com/cloudpulse-runner:prod"

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}

# observability stack for prod
# includes alarms, dashboards, and wiring to ecs + lambda + dynamodb
module "observability" {
  source = "../../modules/observability"

  env                  = "prod"
  ecs_cluster_name     = module.ecs_api.ecs_cluster_name
  ecs_service_name     = module.ecs_api.ecs_service_name
  runner_function_name = module.runner.lambda_function_name
  dynamodb_table_name  = module.results_table.table_name

  alarm_actions = [] # can attach sns/slack in future
  ok_actions    = []

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}
