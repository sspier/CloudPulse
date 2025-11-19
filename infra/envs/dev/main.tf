terraform {
  required_version = ">= 1.11.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket       = "cloudpulse-tf-state"
    key          = "envs/dev/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = "us-east-1"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_ecr_repository" "api" {
  name = "cloudpulse-api"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }
}

locals {
  api_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_repository.api.name}:dev"
}


module "vpc" {
  source = "../../modules/vpc"

  vpc_cidr            = "10.0.0.0/16"
  public_subnet_cidrs = ["10.0.0.0/24", "10.0.1.0/24"]
  private_subnet_cidrs = [
    "10.0.2.0/24",
    "10.0.3.0/24",
  ]

  tags = {
    Project = "cloudpulse"
    Env     = "dev"
  }

}

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

