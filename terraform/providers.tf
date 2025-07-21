terraform {
  backend "s3" {
    bucket = "steve-rhoton-tfstate"
    key    = "sr-labor-line-sandbox"
    region = "us-west-2"
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