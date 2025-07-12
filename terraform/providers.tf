terraform {
  backend "s3" {
    bucket = "srhoton-tfstate"
    key    = "steverhoton-labor-lines"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
      Repository  = "steverhoton-labor-lines"
    }
  }
}