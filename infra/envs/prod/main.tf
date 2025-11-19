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
    key          = "envs/prod/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = "us-east-1"
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
    Env     = "prod"
  }
}

module "results_table" {
  source = "../../modules/dynamodb_results"

  table_name_prefix = "cloudpulse-probe-results"
  env               = "prod"

  tags = {
    Project = "cloudpulse"
    Env     = "prod"
  }
}

