variable "prefix" { default = "" }
variable "sg_name" { default = "" }
variable "local_public_ip" { default = "" }

locals {
  is_networking = var.module_name == "networking"
}

data "aws_vpc" "default" {
  count   = local.is_networking ? 1 : 0
  default = true
}

data "aws_subnets" "default" {
  count = local.is_networking ? 1 : 0
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default[0].id]
  }
}

resource "aws_security_group" "ecs" {
  count       = local.is_networking ? 1 : 0
  name        = var.sg_name
  description = "ECS GPU cluster security group"
  vpc_id      = data.aws_vpc.default[0].id
  tags        = merge(var.tags, { Name = var.sg_name })

  # Full access from local IP
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["${var.local_public_ip}/32"]
    description = "Full access from local IP"
  }

  # Self-referential for cluster comms
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
    description = "Cluster internal"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All egress"
  }
}

output "vpc_id" {
  value = local.is_networking ? data.aws_vpc.default[0].id : ""
}

output "subnet_ids" {
  value = local.is_networking ? data.aws_subnets.default[0].ids : []
}

output "security_group_id" {
  value = local.is_networking ? aws_security_group.ecs[0].id : ""
}