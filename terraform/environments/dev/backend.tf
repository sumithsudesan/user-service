terraform {
  backend "s3" {
    bucket         = "user-platform-tfstate"
    key            = "dev/terraform.tfstate"
    region         = "eu-west-1"
    use_lockfile   = true
    encrypt        = true
  }
}
