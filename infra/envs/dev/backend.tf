terraform {
  # store terraform state in s3 so multiple environments and users can share state safely
  # locking prevents concurrent applies from corrupting state
  backend "s3" {
    bucket       = "cloudpulse-tf-state"
    key          = "envs/dev/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true # encrypt state at rest
    use_lockfile = true # enable native s3 state locking
  }
}
