terraform {
  # store production state in s3 with locking enabled
  # prod uses its own tfstate file to keep environments isolated
  backend "s3" {
    bucket       = "cloudpulse-tf-state"
    key          = "envs/prod/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true # enable native s3 state locking
  }
}
