variable "project_name" {
  description = "Project name used as prefix for all resources"
  type        = string
  default     = "InferenceProfiler"
}

variable "region" {
  description = "The AWS region to deploy resources in"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type for GPU nodes"
  type        = string
  default     = "g4dn.xlarge"
}

variable "desired_capacity" {
  description = "Desired number of GPU instances"
  type        = number
  default     = 1
}

locals {
  prefix = var.project_name
  tags   = { Project = var.project_name }

  iam_role_name     = "${local.prefix}-ecs-role"
  profile_name      = "${local.prefix}-profile"
  cluster_name      = "${local.prefix}-gpu-cluster"
  capacity_provider = "${local.prefix}-gpu-cp"
  sg_name           = "${local.prefix}-sg"
  asg_name          = "${local.prefix}-asg"
  launch_template   = "${local.prefix}-lt"
}