variable "cluster_name" { default = "" }
variable "capacity_provider_name" { default = "" }
variable "asg_name" { default = "" }
variable "launch_template_name" { default = "" }
variable "instance_type" { default = "" }
variable "desired_capacity" { default = 1 }
variable "instance_profile_arn" { default = "" }
variable "vpc_id" { default = "" }
variable "subnet_ids" { default = [] }
variable "security_group_id" { default = "" }

locals {
  is_ecs = var.module_name == "ecs_cluster"
}

data "aws_ssm_parameter" "ecs_gpu_ami" {
  count = local.is_ecs ? 1 : 0
  name  = "/aws/service/ecs/optimized-ami/amazon-linux-2/gpu/recommended/image_id"
}

resource "aws_ecs_cluster" "main" {
  count = local.is_ecs ? 1 : 0
  name  = var.cluster_name
  tags  = var.tags

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_launch_template" "gpu" {
  count         = local.is_ecs ? 1 : 0
  name          = var.launch_template_name
  image_id      = data.aws_ssm_parameter.ecs_gpu_ami[0].value
  instance_type = var.instance_type
  tags          = var.tags

  iam_instance_profile { arn = var.instance_profile_arn }
  vpc_security_group_ids = [var.security_group_id]

  user_data = base64encode(<<-EOF
    #!/bin/bash
    echo "ECS_CLUSTER=${var.cluster_name}" >> /etc/ecs/ecs.config
    echo "ECS_ENABLE_GPU_SUPPORT=true" >> /etc/ecs/ecs.config
  EOF
  )

  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size = 50
      volume_type = "gp3"
    }
  }

  tag_specifications {
    resource_type = "instance"
    tags          = merge(var.tags, { Name = "${var.cluster_name}-gpu-node" })
  }
}

resource "aws_autoscaling_group" "gpu" {
  count               = local.is_ecs ? 1 : 0
  name                = var.asg_name
  desired_capacity    = var.desired_capacity
  max_size            = var.desired_capacity + 1
  min_size            = 0
  vpc_zone_identifier = var.subnet_ids

  launch_template {
    id      = aws_launch_template.gpu[0].id
    version = "$Latest"
  }

  tag {
    key                 = "AmazonECSManaged"
    value               = "true"
    propagate_at_launch = true
  }

  dynamic "tag" {
    for_each = var.tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }

  lifecycle { create_before_destroy = true }
}

resource "aws_ecs_capacity_provider" "gpu" {
  count = local.is_ecs ? 1 : 0
  name  = var.capacity_provider_name
  tags  = var.tags

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.gpu[0].arn
    managed_termination_protection = "DISABLED"

    managed_scaling {
      status                    = "ENABLED"
      target_capacity           = 100
      minimum_scaling_step_size = 1
      maximum_scaling_step_size = 1
    }
  }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  count              = local.is_ecs ? 1 : 0
  cluster_name       = aws_ecs_cluster.main[0].name
  capacity_providers = [aws_ecs_capacity_provider.gpu[0].name]

  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.gpu[0].name
    weight            = 1
  }
}

output "cluster_name" {
  value = local.is_ecs ? aws_ecs_cluster.main[0].name : ""
}

output "cluster_arn" {
  value = local.is_ecs ? aws_ecs_cluster.main[0].arn : ""
}

output "capacity_provider_name" {
  value = local.is_ecs ? aws_ecs_capacity_provider.gpu[0].name : ""
}