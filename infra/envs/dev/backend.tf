terraform {
  backend "s3" {
    bucket = "hesp-landing-zone-tfstate"
    key    = "dev/terraform.tfstate"
    region = "us-east-1"
  }
}
