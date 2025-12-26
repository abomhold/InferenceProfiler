terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

provider "aws" {
  region = var.region
}

# Fetch local public IP
data "http" "local_ip" {
  url = "https://checkip.amazonaws.com"
}

locals {
  local_public_ip = trimspace(data.http.local_ip.response_body)
}

module "identity" {
  source       = "./modules"
  module_name  = "identity"
  role_name    = local.iam_role_name
  profile_name = local.profile_name
  tags         = local.tags
}

module "networking" {
  source          = "./modules"
  module_name     = "networking"
  prefix          = local.prefix
  sg_name         = local.sg_name
  local_public_ip = local.local_public_ip
  tags            = local.tags
}

module "ecs_cluster" {
  source = "./modules"
  module_name = "ecs_cluster"

  cluster_name           = local.cluster_name
  capacity_provider_name = local.capacity_provider
  asg_name               = local.asg_name
  launch_template_name   = local.launch_template
  instance_type          = var.instance_type
  desired_capacity       = var.desired_capacity
  instance_profile_arn   = module.identity.instance_profile_arn
  vpc_id                 = module.networking.vpc_id
  subnet_ids             = module.networking.subnet_ids
  security_group_id      = module.networking.security_group_id
  tags                   = local.tags

  depends_on = [module.identity, module.networking]
}